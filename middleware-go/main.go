package main

import (
	"context"
	"log"
	"net"
	"sort"
	"time"

	pb "agenic-middleware/middleware-go/pb"

	"google.golang.org/grpc"
)

var (
	// 티켓 재고는 단 1장!
	ticketStock = 1
	// 에이전트들의 요청을 임시로 담아둘 대기열(Channel)
	commitChan = make(chan CommitTask, 1000)
	// 2. QCFuse(조회)용 채널 (추가됨!)
	readChan = make(chan ReadTask, 1000)
)

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
	ticker := time.NewTicker(100 * time.Millisecond)
	var batch []ReadTask

	for {
		select {
		case task := <-readChan:
			batch = append(batch, task)

		case <-ticker.C:
			if len(batch) == 0 {
				continue
			}

			// 0.1초 동안 모인 요청들을 1개의 DB I/O로 융합 처리!
			log.Printf("==================================================")
			log.Printf("🔍 [QCFuse 작동] %d명의 조회 요청을 1회의 DB I/O로 병합 처리!", len(batch))
			log.Printf("==================================================\n")

			// 단 1회의 DB 조회 결과라고 가정
			fusedResponse := &pb.ReadResponse{
				Success: true,
				Data:    "재고 1장 남음",
				Message: "QCFuse 시맨틱 캐시 적중",
			}

			// 모인 수십/수백 명의 에이전트에게 동시에 결과 브로드캐스트 (Broadcast)
			for _, task := range batch {
				task.Res <- fusedResponse
			}

			// 큐 초기화
			batch = nil
		}
	}
}

// [Phase 3] 커밋 요청 수신부
func (s *server) CommitTransaction(ctx context.Context, req *pb.CommitRequest) (*pb.CommitResponse, error) {
	resChan := make(chan *pb.CommitResponse)

	// 요청을 받으면 즉시 처리하지 않고 ATCC 스케줄러 큐(Queue)로 넘겨버려!
	commitChan <- CommitTask{Req: req, Res: resChan}

	// 스케줄러가 심사를 마치고 결과를 돌려줄 때까지 대기
	res := <-resChan
	return res, nil
}

// 👑 [핵심] ATCC 지능형 스케줄러 (백그라운드 실행)
func ATCCScheduler() {
	// 3초 단위(윈도우)로 모인 요청들을 한 번에 심사해.
	ticker := time.NewTicker(3 * time.Second)
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

			// 이미 누군가 티켓을 사갔다면(재고 0), 무조건 롤백 처리
			if ticketStock <= 0 {
				for _, task := range batch {
					task.Res <- &pb.CommitResponse{
						Success: false, IsRolledBack: true, Message: "재고 소진으로 인한 롤백", SavedCostUsd: 0.0,
					}
				}
				batch = nil
				continue
			}

			// 🚀 1. 매몰 비용(Sunk Cost) 계산 및 내림차순 정렬
			sort.Slice(batch, func(i, j int) bool {
				costI := calculateSunkCost(batch[i].Req)
				costJ := calculateSunkCost(batch[j].Req)
				return costI > costJ // 비용이 큰 사람이 배열의 맨 앞으로(1등) 옴
			})

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
			winnerCost := calculateSunkCost(winner.Req)
			log.Printf("==================================================")
			log.Printf("🏆 [ATCC 승인] %s (투자 비용: $%.2f) - 재고 획득!", winner.Req.GetAgentId(), winnerCost)

			winner.Res <- &pb.CommitResponse{
				Success: true, IsRolledBack: false, Message: "최고 비용 에이전트 커밋 성공",
			}

			// ⛔ 3. 나머지 에이전트 롤백 처리 및 방어 비용 산출
			for i := 1; i < len(batch); i++ {
				loser := batch[i]
				savedCost := calculateSunkCost(loser.Req)
				log.Printf("⛔ [ATCC 롤백] %s (방어한 손실 비용: $%.2f)", loser.Req.GetAgentId(), savedCost)

				loser.Res <- &pb.CommitResponse{
					Success: false, IsRolledBack: true, Message: "경합 패배", SavedCostUsd: savedCost,
				}
			}
			log.Printf("==================================================\n")

			// 큐 초기화 (다음 3초를 위해)
			batch = nil

			ticketStock = 1
		}
	}
}

// 💰 매몰 비용 계산 공식 (토큰 가중치 + 시간 가중치)
func calculateSunkCost(req *pb.CommitRequest) float32 {
	tokens := float32(req.GetAccumulatedTokens())
	latency := float32(req.GetInferenceLatencySec())

	// 예시: 1토큰 = $0.002, 1초 대기 = $0.5 로 가정하여 달러($) 단위로 환산
	return (tokens * 0.002) + (latency * 0.5)
}

func main() {
	// 1. 두 개의 핵심 두뇌(스케줄러)를 백그라운드 엔진으로 가동!
	go QCFuseScheduler()
	go ATCCScheduler()

	// 2. gRPC 서버 구동
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("포트 리스닝 실패: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterTransactionMiddlewareServer(s, &server{})

	log.Printf("🚀 미들웨어 서버가 [::]:50051 포트에서 실행 중입니다...")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("서버 실행 실패: %v", err)
	}
}
