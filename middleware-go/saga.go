package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"sync"
	"time"

	pb "agenic-middleware/middleware-go/pb"
)

const (
	SagaStatusActive           = "ACTIVE"
	SagaStatusValidated        = "VALIDATED"
	SagaStatusValidationFailed = "VALIDATION_FAILED"
	SagaStatusCommitted        = "COMMITTED"
	SagaStatusAborted          = "ABORTED"
	SagaStatusCompensated      = "COMPENSATED"

	SagaStepStatusCompleted   = "COMPLETED"
	SagaStepStatusCompensated = "COMPENSATED"
)

type SagaStepRecord struct {
	ID                 string
	Action             string
	Result             string
	CompensationAction string
	Status             string
}

type SagaRecord struct {
	ID                string
	AgentID           string
	Goal              string
	Status            string
	Steps             []SagaStepRecord
	ValidationMessage string
	AbortReason       string
	CompensationLog   []string
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

type SagaCoordinator struct {
	mu    sync.RWMutex
	sagas map[string]*SagaRecord
}

func NewSagaCoordinator() *SagaCoordinator {
	return &SagaCoordinator{sagas: make(map[string]*SagaRecord)}
}

func (c *SagaCoordinator) Begin(agentID, goal string) SagaRecord {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	saga := &SagaRecord{
		ID:        newSagaID(),
		AgentID:   agentID,
		Goal:      goal,
		Status:    SagaStatusActive,
		CreatedAt: now,
		UpdatedAt: now,
	}
	c.sagas[saga.ID] = saga
	return cloneSaga(saga)
}

func (c *SagaCoordinator) RegisterStep(sagaID, stepID, action, result, compensationAction string) (SagaRecord, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	saga, ok := c.sagas[sagaID]
	if !ok {
		return SagaRecord{}, fmt.Errorf("saga %q not found", sagaID)
	}
	if saga.Status != SagaStatusActive && saga.Status != SagaStatusValidationFailed {
		return cloneSaga(saga), fmt.Errorf("cannot register step while saga is %s", saga.Status)
	}
	if stepID == "" || action == "" {
		return cloneSaga(saga), errors.New("step_id and action are required")
	}
	for _, step := range saga.Steps {
		if step.ID == stepID {
			return cloneSaga(saga), fmt.Errorf("step %q already exists", stepID)
		}
	}

	saga.Steps = append(saga.Steps, SagaStepRecord{
		ID:                 stepID,
		Action:             action,
		Result:             result,
		CompensationAction: compensationAction,
		Status:             SagaStepStatusCompleted,
	})
	saga.Status = SagaStatusActive
	saga.ValidationMessage = ""
	saga.UpdatedAt = time.Now()
	return cloneSaga(saga), nil
}

func (c *SagaCoordinator) Validate(sagaID string) (SagaRecord, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	saga, ok := c.sagas[sagaID]
	if !ok {
		return SagaRecord{}, fmt.Errorf("saga %q not found", sagaID)
	}
	if saga.Status != SagaStatusActive && saga.Status != SagaStatusValidationFailed {
		return cloneSaga(saga), fmt.Errorf("cannot validate saga while status is %s", saga.Status)
	}
	if len(saga.Steps) == 0 {
		saga.Status = SagaStatusValidationFailed
		saga.ValidationMessage = "최소 한 개 이상의 완료된 step이 필요합니다"
		saga.UpdatedAt = time.Now()
		return cloneSaga(saga), errors.New(saga.ValidationMessage)
	}
	for _, step := range saga.Steps {
		if step.Status != SagaStepStatusCompleted {
			saga.Status = SagaStatusValidationFailed
			saga.ValidationMessage = fmt.Sprintf("step %s is not completed", step.ID)
			saga.UpdatedAt = time.Now()
			return cloneSaga(saga), errors.New(saga.ValidationMessage)
		}
	}

	saga.Status = SagaStatusValidated
	saga.ValidationMessage = "deterministic validation 통과"
	saga.UpdatedAt = time.Now()
	return cloneSaga(saga), nil
}

func (c *SagaCoordinator) Commit(sagaID string) (SagaRecord, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	saga, ok := c.sagas[sagaID]
	if !ok {
		return SagaRecord{}, fmt.Errorf("saga %q not found", sagaID)
	}
	if saga.Status != SagaStatusValidated {
		return cloneSaga(saga), fmt.Errorf("saga must be validated before commit, current status is %s", saga.Status)
	}
	saga.Status = SagaStatusCommitted
	saga.UpdatedAt = time.Now()
	return cloneSaga(saga), nil
}

func (c *SagaCoordinator) Abort(sagaID, reason string) (SagaRecord, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	saga, ok := c.sagas[sagaID]
	if !ok {
		return SagaRecord{}, fmt.Errorf("saga %q not found", sagaID)
	}
	if saga.Status == SagaStatusCommitted {
		return cloneSaga(saga), errors.New("committed saga cannot be aborted")
	}

	saga.Status = SagaStatusAborted
	saga.AbortReason = reason
	saga.CompensationLog = nil

	for i := len(saga.Steps) - 1; i >= 0; i-- {
		step := &saga.Steps[i]
		if step.Status != SagaStepStatusCompleted {
			continue
		}
		if step.CompensationAction != "" {
			saga.CompensationLog = append(saga.CompensationLog, step.CompensationAction)
		}
		step.Status = SagaStepStatusCompensated
	}
	saga.Status = SagaStatusCompensated
	saga.UpdatedAt = time.Now()
	return cloneSaga(saga), nil
}

func (c *SagaCoordinator) Get(sagaID string) (SagaRecord, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	saga, ok := c.sagas[sagaID]
	if !ok {
		return SagaRecord{}, fmt.Errorf("saga %q not found", sagaID)
	}
	return cloneSaga(saga), nil
}

func (c *SagaCoordinator) Reset() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.sagas = make(map[string]*SagaRecord)
}

func cloneSaga(saga *SagaRecord) SagaRecord {
	cloned := *saga
	cloned.Steps = append([]SagaStepRecord(nil), saga.Steps...)
	cloned.CompensationLog = append([]string(nil), saga.CompensationLog...)
	return cloned
}

func newSagaID() string {
	raw := make([]byte, 8)
	if _, err := rand.Read(raw); err != nil {
		return fmt.Sprintf("saga-%d", time.Now().UnixNano())
	}
	return "saga-" + hex.EncodeToString(raw)
}

func sagaToProto(saga SagaRecord) *pb.SagaState {
	steps := make([]*pb.SagaStep, 0, len(saga.Steps))
	for _, step := range saga.Steps {
		steps = append(steps, &pb.SagaStep{
			StepId:             step.ID,
			Action:             step.Action,
			Result:             step.Result,
			CompensationAction: step.CompensationAction,
			Status:             step.Status,
		})
	}
	return &pb.SagaState{
		SagaId:               saga.ID,
		AgentId:              saga.AgentID,
		Goal:                 saga.Goal,
		Status:               saga.Status,
		Steps:                steps,
		ValidationMessage:    saga.ValidationMessage,
		AbortReason:          saga.AbortReason,
		CreatedAtUnixSeconds: saga.CreatedAt.Unix(),
		UpdatedAtUnixSeconds: saga.UpdatedAt.Unix(),
	}
}

func sagaResponse(saga SagaRecord, message string) *pb.SagaResponse {
	return &pb.SagaResponse{Success: true, Message: message, Saga: sagaToProto(saga)}
}

func sagaErrorResponse(saga SagaRecord, err error) *pb.SagaResponse {
	response := &pb.SagaResponse{Success: false, Message: err.Error()}
	if saga.ID != "" {
		response.Saga = sagaToProto(saga)
	}
	return response
}

func (s *server) BeginSaga(_ context.Context, req *pb.BeginSagaRequest) (*pb.SagaResponse, error) {
	if req.GetAgentId() == "" || req.GetGoal() == "" {
		return &pb.SagaResponse{Success: false, Message: "agent_id and goal are required"}, nil
	}
	saga := sagaCoordinator.Begin(req.GetAgentId(), req.GetGoal())
	metrics.RecordSagaStarted()
	return sagaResponse(saga, "Saga workflow 시작"), nil
}

func (s *server) RegisterSagaStep(_ context.Context, req *pb.RegisterSagaStepRequest) (*pb.SagaResponse, error) {
	saga, err := sagaCoordinator.RegisterStep(
		req.GetSagaId(),
		req.GetStepId(),
		req.GetAction(),
		req.GetResult(),
		req.GetCompensationAction(),
	)
	if err != nil {
		return sagaErrorResponse(saga, err), nil
	}
	metrics.RecordSagaStep()
	return sagaResponse(saga, "Saga checkpoint 등록"), nil
}

func (s *server) ValidateSaga(_ context.Context, req *pb.ValidateSagaRequest) (*pb.SagaResponse, error) {
	saga, err := sagaCoordinator.Validate(req.GetSagaId())
	if err != nil {
		metrics.RecordSagaValidationFailure()
		return sagaErrorResponse(saga, err), nil
	}
	metrics.RecordSagaValidated()
	return sagaResponse(saga, "Saga deterministic validation 통과"), nil
}

func (s *server) AbortSaga(_ context.Context, req *pb.AbortSagaRequest) (*pb.SagaResponse, error) {
	saga, err := sagaCoordinator.Abort(req.GetSagaId(), req.GetReason())
	if err != nil {
		return sagaErrorResponse(saga, err), nil
	}
	metrics.RecordSagaCompensated(len(saga.CompensationLog))
	return sagaResponse(saga, "Saga abort 및 compensation 완료"), nil
}

func (s *server) GetSagaState(_ context.Context, req *pb.GetSagaStateRequest) (*pb.SagaState, error) {
	saga, err := sagaCoordinator.Get(req.GetSagaId())
	if err != nil {
		return &pb.SagaState{SagaId: req.GetSagaId(), Status: "NOT_FOUND", ValidationMessage: err.Error()}, nil
	}
	return sagaToProto(saga), nil
}
