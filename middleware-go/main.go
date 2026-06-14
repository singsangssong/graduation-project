package main

import (
	"context"
	"encoding/json"
	"log"
	"net"
	"net/http"
	"os"
	"sort"
	"strconv"
	"sync"
	"time"

	pb "agenic-middleware/middleware-go/pb"

	"google.golang.org/grpc"
)

var (
	// 티켓 재고는 단 1장!
	ticketStock   = 1
	ticketStockMu sync.Mutex
	// 에이전트들의 요청을 임시로 담아둘 대기열(Channel)
	commitChan = make(chan CommitTask, 1000)
	// 2. QCFuse(조회)용 채널 (추가됨!)
	readChan = make(chan ReadTask, 1000)

	appConfig       = loadConfig()
	metrics         = NewMiddlewareMetrics()
	sagaCoordinator = NewSagaCoordinator()
)

type Config struct {
	GrpcAddr           string  `json:"grpc_addr"`
	MetricsAddr        string  `json:"metrics_addr"`
	QCFuseWindowMs     int     `json:"qcfuse_window_ms"`
	ATCCWindowMs       int     `json:"atcc_window_ms"`
	TokenCostWeight    float32 `json:"token_cost_weight"`
	LatencyCostWeight  float32 `json:"latency_cost_weight"`
	InitialTicketStock int     `json:"initial_ticket_stock"`
	SagaDBPath         string  `json:"saga_db_path"`
}

func loadConfig() Config {
	return Config{
		GrpcAddr:           getEnv("MIDDLEWARE_GRPC_ADDR", ":50051"),
		MetricsAddr:        getEnv("MIDDLEWARE_METRICS_ADDR", ":8080"),
		QCFuseWindowMs:     getEnvInt("QCFUSE_WINDOW_MS", 100),
		ATCCWindowMs:       getEnvInt("ATCC_WINDOW_MS", 3000),
		TokenCostWeight:    getEnvFloat32("ATCC_TOKEN_WEIGHT", 0.002),
		LatencyCostWeight:  getEnvFloat32("ATCC_LATENCY_WEIGHT", 0.5),
		InitialTicketStock: getEnvInt("TICKET_STOCK", 1),
		SagaDBPath:         getEnv("SAGA_DB_PATH", "data/middleware.db"),
	}
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func getEnvInt(key string, defaultValue int) int {
	raw := os.Getenv(key)
	if raw == "" {
		return defaultValue
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		log.Printf("[config] %s=%q 파싱 실패, 기본값 %d 사용", key, raw, defaultValue)
		return defaultValue
	}
	return value
}

func getEnvFloat32(key string, defaultValue float32) float32 {
	raw := os.Getenv(key)
	if raw == "" {
		return defaultValue
	}
	value, err := strconv.ParseFloat(raw, 32)
	if err != nil {
		log.Printf("[config] %s=%q 파싱 실패, 기본값 %.4f 사용", key, raw, defaultValue)
		return defaultValue
	}
	return float32(value)
}

type MiddlewareMetrics struct {
	mu                     *sync.Mutex
	ReadRequests           int     `json:"read_requests"`
	FusedReadBatches       int     `json:"fused_read_batches"`
	FusedReadRequests      int     `json:"fused_read_requests"`
	SavedDBReads           int     `json:"saved_db_reads"`
	CommitRequests         int     `json:"commit_requests"`
	CommitBatches          int     `json:"commit_batches"`
	ApprovedCommits        int     `json:"approved_commits"`
	RolledBackCommits      int     `json:"rolled_back_commits"`
	TotalSavedCostUsd      float32 `json:"total_saved_cost_usd"`
	LastReadBatchSize      int     `json:"last_read_batch_size"`
	LastCommitBatchSize    int     `json:"last_commit_batch_size"`
	LastWinnerAgentID      string  `json:"last_winner_agent_id"`
	LastWinnerCostUsd      float32 `json:"last_winner_cost_usd"`
	LastUpdatedUnixSeconds int64   `json:"last_updated_unix_seconds"`
	SagasStarted           int     `json:"sagas_started"`
	SagaStepsRegistered    int     `json:"saga_steps_registered"`
	SagasValidated         int     `json:"sagas_validated"`
	SagaValidationFailures int     `json:"saga_validation_failures"`
	SagasCompensated       int     `json:"sagas_compensated"`
	CompensationActions    int     `json:"compensation_actions"`
}

func NewMiddlewareMetrics() *MiddlewareMetrics {
	return &MiddlewareMetrics{mu: &sync.Mutex{}}
}

func (m *MiddlewareMetrics) Snapshot() MiddlewareMetrics {
	m.mu.Lock()
	defer m.mu.Unlock()
	return *m
}

func (m *MiddlewareMetrics) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.ReadRequests = 0
	m.FusedReadBatches = 0
	m.FusedReadRequests = 0
	m.SavedDBReads = 0
	m.CommitRequests = 0
	m.CommitBatches = 0
	m.ApprovedCommits = 0
	m.RolledBackCommits = 0
	m.TotalSavedCostUsd = 0
	m.LastReadBatchSize = 0
	m.LastCommitBatchSize = 0
	m.LastWinnerAgentID = ""
	m.LastWinnerCostUsd = 0
	m.LastUpdatedUnixSeconds = time.Now().Unix()
	m.SagasStarted = 0
	m.SagaStepsRegistered = 0
	m.SagasValidated = 0
	m.SagaValidationFailures = 0
	m.SagasCompensated = 0
	m.CompensationActions = 0
}

func (m *MiddlewareMetrics) RecordReadRequest() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.ReadRequests++
	m.LastUpdatedUnixSeconds = time.Now().Unix()
}

func (m *MiddlewareMetrics) RecordFusedReadBatch(size int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.FusedReadBatches++
	m.FusedReadRequests += size
	m.LastReadBatchSize = size
	if size > 1 {
		m.SavedDBReads += size - 1
	}
	m.LastUpdatedUnixSeconds = time.Now().Unix()
}

func (m *MiddlewareMetrics) RecordCommitRequest() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.CommitRequests++
	m.LastUpdatedUnixSeconds = time.Now().Unix()
}

func (m *MiddlewareMetrics) RecordCommitBatch(size int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.CommitBatches++
	m.LastCommitBatchSize = size
	m.LastUpdatedUnixSeconds = time.Now().Unix()
}

func (m *MiddlewareMetrics) RecordApprovedCommit(agentID string, cost float32) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.ApprovedCommits++
	m.LastWinnerAgentID = agentID
	m.LastWinnerCostUsd = cost
	m.LastUpdatedUnixSeconds = time.Now().Unix()
}

func (m *MiddlewareMetrics) RecordRolledBackCommit(savedCost float32) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.RolledBackCommits++
	m.TotalSavedCostUsd += savedCost
	m.LastUpdatedUnixSeconds = time.Now().Unix()
}

func (m *MiddlewareMetrics) RecordSagaStarted() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.SagasStarted++
	m.LastUpdatedUnixSeconds = time.Now().Unix()
}

func (m *MiddlewareMetrics) RecordSagaStep() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.SagaStepsRegistered++
	m.LastUpdatedUnixSeconds = time.Now().Unix()
}

func (m *MiddlewareMetrics) RecordSagaValidated() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.SagasValidated++
	m.LastUpdatedUnixSeconds = time.Now().Unix()
}

func (m *MiddlewareMetrics) RecordSagaValidationFailure() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.SagaValidationFailures++
	m.LastUpdatedUnixSeconds = time.Now().Unix()
}

func (m *MiddlewareMetrics) RecordSagaCompensated(actions int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.SagasCompensated++
	m.CompensationActions += actions
	m.LastUpdatedUnixSeconds = time.Now().Unix()
}

type MetricsResponse struct {
	Config  Config            `json:"config"`
	Metrics MiddlewareMetrics `json:"metrics"`
}

// CommitTask는 에이전트의 요청과 응답을 돌려줄 통로를 묶어둔 구조체야.
type CommitTask struct {
	Req *pb.CommitRequest
	Res chan *pb.CommitResponse
}

// QCFuse 처리를 위한 조회 태스크 구조체
type ReadTask struct {
	Req *pb.ReadRequest
	Res chan *pb.ReadResponse
}

type server struct {
	pb.UnimplementedTransactionMiddlewareServer
}

// =====================================================================
// [Phase 1] ReadResource: 에이전트의 조회 요청 수신부
// =====================================================================
func (s *server) ReadResource(ctx context.Context, req *pb.ReadRequest) (*pb.ReadResponse, error) {
	resChan := make(chan *pb.ReadResponse)
	metrics.RecordReadRequest()

	// 들어온 조회 요청을 즉시 DB로 보내지 않고, QCFuse 채널(바구니)에 담음!
	readChan <- ReadTask{Req: req, Res: resChan}

	// QCFuse 스케줄러가 융합(Fusion) 처리를 마칠 때까지 대기
	res := <-resChan
	return res, nil
}

// =====================================================================
// ⚡ [핵심 1] QCFuse 캐시 융합 스케줄러 (백그라운드 실행)
// =====================================================================
func QCFuseScheduler() {
	// 0.1초(100ms) 단위의 초고속 스케줄링 윈도우
	ticker := time.NewTicker(time.Duration(appConfig.QCFuseWindowMs) * time.Millisecond)
	var batch []ReadTask

	for {
		select {
		case task := <-readChan:
			batch = append(batch, task)

		case <-ticker.C:
			if len(batch) == 0 {
				continue
			}

			// 서로 다른 resource/intent 요청은 섞지 않고 key별로 fusion한다.
			for key, tasks := range groupReadTasksByResource(batch) {
				metrics.RecordFusedReadBatch(len(tasks))
				log.Printf("🔍 [QCFuse 작동] key=%s, %d개 요청을 1회의 logical DB I/O로 병합", key, len(tasks))

				fusedResponse := &pb.ReadResponse{
					Success: true,
					Data:    tasks[0].Req.GetResourceId() + " 재고 1장 남음",
					Message: "QCFuse resource-aware read fusion 적중",
				}
				for _, task := range tasks {
					task.Res <- fusedResponse
				}
			}

			// 큐 초기화
			batch = nil
		}
	}
}

// [Phase 3] 커밋 요청 수신부
func (s *server) CommitTransaction(ctx context.Context, req *pb.CommitRequest) (*pb.CommitResponse, error) {
	if req.GetSagaId() != "" {
		saga, err := sagaCoordinator.Get(req.GetSagaId())
		if err != nil {
			return &pb.CommitResponse{Success: false, IsRolledBack: true, Message: err.Error(), SagaStatus: "NOT_FOUND"}, nil
		}
		if saga.Status != SagaStatusValidated {
			return &pb.CommitResponse{
				Success: false, IsRolledBack: true,
				Message: "Saga validation이 commit 전에 필요합니다", SagaStatus: saga.Status,
			}, nil
		}
	}

	resChan := make(chan *pb.CommitResponse)
	metrics.RecordCommitRequest()

	// 요청을 받으면 즉시 처리하지 않고 ATCC 스케줄러 큐(Queue)로 넘겨버려!
	commitChan <- CommitTask{Req: req, Res: resChan}

	// 스케줄러가 심사를 마치고 결과를 돌려줄 때까지 대기
	res := <-resChan
	return res, nil
}

// 👑 [핵심] ATCC 지능형 스케줄러 (백그라운드 실행)
func ATCCScheduler() {
	// 3초 단위(윈도우)로 모인 요청들을 한 번에 심사해.
	ticker := time.NewTicker(time.Duration(appConfig.ATCCWindowMs) * time.Millisecond)
	var batch []CommitTask

	for {
		select {
		case task := <-commitChan:
			// 에이전트 요청이 오면 일단 배열(batch)에 차곡차곡 모아둠
			batch = append(batch, task)

		case <-ticker.C:
			// 3초가 지나고 심사 시작!
			if len(batch) == 0 {
				continue
			}
			metrics.RecordCommitBatch(len(batch))

			ticketStockMu.Lock()
			// 이미 누군가 티켓을 사갔다면(재고 0), 무조건 롤백 처리
			if ticketStock <= 0 {
				ticketStockMu.Unlock()
				for _, task := range batch {
					metrics.RecordRolledBackCommit(0)
					task.Res <- &pb.CommitResponse{
						Success: false, IsRolledBack: true, Message: "재고 소진으로 인한 롤백", SavedCostUsd: 0.0,
					}
				}
				batch = nil
				continue
			}

			// 🚀 1. 매몰 비용(Sunk Cost) 계산 및 내림차순 정렬
			batch = rankCommitTasks(batch, appConfig)

			// 📊 [추가된 코드] Top 10 리더보드 출력
			log.Printf("==================================================")
			log.Printf("📊 [ATCC 경합 리더보드 - 총 %d명 경합]", len(batch))
			limit := len(batch)
			if limit > 10 {
				limit = 10 // 상위 10명만 출력
			}
			for i := 0; i < limit; i++ {
				cost := calculateSunkCost(batch[i].Req)
				log.Printf("  %d등: %s (매몰 비용: $%.2f)", i+1, batch[i].Req.GetAgentId(), cost)
			}
			log.Printf("--------------------------------------------------")

			// 🏆 2. 1등 에이전트 승인 (재고 차감)
			winner := batch[0]
			ticketStock--
			ticketStockMu.Unlock()
			winnerCost := calculateSunkCost(winner.Req)
			winnerSagaStatus := ""
			if winner.Req.GetSagaId() != "" {
				saga, err := sagaCoordinator.Commit(winner.Req.GetSagaId())
				if err != nil {
					log.Printf("[Saga commit 실패] %v", err)
				} else {
					winnerSagaStatus = saga.Status
				}
			}
			metrics.RecordApprovedCommit(winner.Req.GetAgentId(), winnerCost)
			log.Printf("==================================================")
			log.Printf("🏆 [ATCC 승인] %s (투자 비용: $%.2f) - 재고 획득!", winner.Req.GetAgentId(), winnerCost)

			winner.Res <- &pb.CommitResponse{
				Success: true, IsRolledBack: false, Message: "최고 비용 에이전트 커밋 성공", SagaStatus: winnerSagaStatus,
			}

			// ⛔ 3. 나머지 에이전트 롤백 처리 및 방어 비용 산출
			for i := 1; i < len(batch); i++ {
				loser := batch[i]
				savedCost := calculateSunkCost(loser.Req)
				loserSagaStatus := ""
				if loser.Req.GetSagaId() != "" {
					saga, err := sagaCoordinator.Abort(loser.Req.GetSagaId(), "ATCC commit arbitration conflict")
					if err != nil {
						log.Printf("[Saga compensation 실패] %v", err)
					} else {
						loserSagaStatus = saga.Status
						metrics.RecordSagaCompensated(len(saga.CompensationLog))
					}
				}
				metrics.RecordRolledBackCommit(savedCost)
				log.Printf("⛔ [ATCC 롤백] %s (방어한 손실 비용: $%.2f)", loser.Req.GetAgentId(), savedCost)

				loser.Res <- &pb.CommitResponse{
					Success: false, IsRolledBack: true, Message: "경합 패배 및 Saga compensation", SavedCostUsd: savedCost, SagaStatus: loserSagaStatus,
				}
			}
			log.Printf("==================================================\n")

			// 큐 초기화 (다음 3초를 위해)
			batch = nil

			ticketStockMu.Lock()
			ticketStock = appConfig.InitialTicketStock
			ticketStockMu.Unlock()
		}
	}
}

func readFusionKey(req *pb.ReadRequest) string {
	return req.GetResourceId() + "\x00" + req.GetIntent()
}

func groupReadTasksByResource(tasks []ReadTask) map[string][]ReadTask {
	groups := make(map[string][]ReadTask)
	for _, task := range tasks {
		key := readFusionKey(task.Req)
		groups[key] = append(groups[key], task)
	}
	return groups
}

// 💰 매몰 비용 계산 공식 (토큰 가중치 + 시간 가중치)
func calculateSunkCost(req *pb.CommitRequest) float32 {
	return calculateSunkCostWithConfig(req, appConfig)
}

func calculateSunkCostWithConfig(req *pb.CommitRequest, config Config) float32 {
	tokens := float32(req.GetAccumulatedTokens())
	latency := float32(req.GetInferenceLatencySec())

	return (tokens * config.TokenCostWeight) + (latency * config.LatencyCostWeight)
}

func rankCommitTasks(batch []CommitTask, config Config) []CommitTask {
	ranked := append([]CommitTask(nil), batch...)
	sort.Slice(ranked, func(i, j int) bool {
		costI := calculateSunkCostWithConfig(ranked[i].Req, config)
		costJ := calculateSunkCostWithConfig(ranked[j].Req, config)
		return costI > costJ
	})
	return ranked
}

func StartMetricsServer(addr string) {
	mux := http.NewServeMux()
	mux.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(MetricsResponse{
			Config:  appConfig,
			Metrics: metrics.Snapshot(),
		})
	})
	mux.HandleFunc("/reset", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		if r.Method != http.MethodPost {
			http.Error(w, "POST only", http.StatusMethodNotAllowed)
			return
		}
		metrics.Reset()
		sagaCoordinator.Reset()
		ticketStockMu.Lock()
		ticketStock = appConfig.InitialTicketStock
		ticketStockMu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "reset"})
	})
	mux.HandleFunc("/events", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Content-Type", "application/json")
		events, err := sagaCoordinator.Events(r.URL.Query().Get("saga_id"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		_ = json.NewEncoder(w).Encode(events)
	})
	mux.HandleFunc("/resource", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Content-Type", "application/json")
		resourceID := r.URL.Query().Get("resource_id")
		stock, err := sagaCoordinator.ResourceStock(resourceID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"resource_id": resourceID, "available_stock": stock})
	})

	log.Printf("📈 메트릭 서버가 %s 포트에서 실행 중입니다...", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("메트릭 서버 실행 실패: %v", err)
	}
}

func main() {
	persistentCoordinator, err := NewPersistentSagaCoordinator(appConfig.SagaDBPath, appConfig.InitialTicketStock)
	if err != nil {
		log.Fatalf("Saga SQLite store 초기화 실패: %v", err)
	}
	sagaCoordinator = persistentCoordinator
	defer sagaCoordinator.Close()

	ticketStockMu.Lock()
	ticketStock = appConfig.InitialTicketStock
	ticketStockMu.Unlock()
	// 1. 두 개의 핵심 두뇌(스케줄러)를 백그라운드 엔진으로 가동!
	go QCFuseScheduler()
	go ATCCScheduler()
	go StartMetricsServer(appConfig.MetricsAddr)

	// 2. gRPC 서버 구동
	lis, err := net.Listen("tcp", appConfig.GrpcAddr)
	if err != nil {
		log.Fatalf("포트 리스닝 실패: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterTransactionMiddlewareServer(s, &server{})

	log.Printf("🚀 미들웨어 서버가 %s 포트에서 실행 중입니다...", appConfig.GrpcAddr)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("서버 실행 실패: %v", err)
	}
}
