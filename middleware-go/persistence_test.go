package main

import (
	"path/filepath"
	"testing"
)

func TestSQLiteSagaStoreRecoversSagaAfterCoordinatorRestart(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "middleware.db")
	coordinator, err := NewPersistentSagaCoordinator(dbPath, 2)
	if err != nil {
		t.Fatalf("NewPersistentSagaCoordinator() error = %v", err)
	}

	saga := coordinator.Begin("Agent-A", "purchase ticket-A")
	if _, err := coordinator.RegisterStep(saga.ID, "reserve", "reserve ticket-A", "reserved", "release ticket-A"); err != nil {
		t.Fatalf("RegisterStep() error = %v", err)
	}
	coordinator.Close()

	recovered, err := NewPersistentSagaCoordinator(dbPath, 2)
	if err != nil {
		t.Fatalf("reopen coordinator error = %v", err)
	}
	defer recovered.Close()

	got, err := recovered.Get(saga.ID)
	if err != nil {
		t.Fatalf("Get() recovered saga error = %v", err)
	}
	if got.ID != saga.ID || len(got.Steps) != 1 {
		t.Fatalf("recovered saga = %+v", got)
	}
}

func TestSQLiteCompensationRestoresReservedResource(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "middleware.db")
	coordinator, err := NewPersistentSagaCoordinator(dbPath, 2)
	if err != nil {
		t.Fatalf("NewPersistentSagaCoordinator() error = %v", err)
	}
	defer coordinator.Close()

	saga := coordinator.Begin("Agent-B", "purchase ticket-A")
	if _, err := coordinator.RegisterStep(saga.ID, "reserve", "reserve ticket-A", "reserved", "release ticket-A"); err != nil {
		t.Fatalf("RegisterStep() error = %v", err)
	}
	if stock, _ := coordinator.ResourceStock("ticket-A"); stock != 1 {
		t.Fatalf("stock after reserve = %d, want 1", stock)
	}

	if _, err := coordinator.Abort(saga.ID, "test abort"); err != nil {
		t.Fatalf("Abort() error = %v", err)
	}
	if stock, _ := coordinator.ResourceStock("ticket-A"); stock != 2 {
		t.Fatalf("stock after compensation = %d, want 2", stock)
	}
}

func TestSQLiteSagaStoreRecordsEventTimeline(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "middleware.db")
	coordinator, err := NewPersistentSagaCoordinator(dbPath, 1)
	if err != nil {
		t.Fatalf("NewPersistentSagaCoordinator() error = %v", err)
	}
	defer coordinator.Close()

	saga := coordinator.Begin("Agent-C", "purchase ticket-A")
	_, _ = coordinator.RegisterStep(saga.ID, "reserve", "reserve ticket-A", "reserved", "release ticket-A")
	_, _ = coordinator.Validate(saga.ID)
	_, _ = coordinator.Abort(saga.ID, "test abort")

	events, err := coordinator.Events(saga.ID)
	if err != nil {
		t.Fatalf("Events() error = %v", err)
	}
	want := []string{"SAGA_STARTED", "STEP_COMPLETED", "VALIDATION_PASSED", "SAGA_ABORTED", "COMPENSATION_COMPLETED", "SAGA_COMPENSATED"}
	if len(events) != len(want) {
		t.Fatalf("event count = %d, want %d: %+v", len(events), len(want), events)
	}
	for i, eventType := range want {
		if events[i].Type != eventType {
			t.Fatalf("event[%d] = %s, want %s", i, events[i].Type, eventType)
		}
	}
}

func TestSQLiteCompensationIsIdempotent(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "middleware.db")
	coordinator, err := NewPersistentSagaCoordinator(dbPath, 2)
	if err != nil {
		t.Fatalf("NewPersistentSagaCoordinator() error = %v", err)
	}
	defer coordinator.Close()

	saga := coordinator.Begin("Agent-D", "purchase ticket-A")
	_, _ = coordinator.RegisterStep(saga.ID, "reserve", "reserve ticket-A", "reserved", "release ticket-A")
	_, _ = coordinator.Abort(saga.ID, "first abort")
	if err := coordinator.store.ExecuteCompensation(saga.ID, "release ticket-A"); err != nil {
		t.Fatalf("second compensation error = %v", err)
	}
	if stock, _ := coordinator.ResourceStock("ticket-A"); stock != 2 {
		t.Fatalf("stock after duplicate compensation = %d, want 2", stock)
	}
}

func TestSQLiteUnsupportedCompensationMarksSagaFailed(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "middleware.db")
	coordinator, err := NewPersistentSagaCoordinator(dbPath, 1)
	if err != nil {
		t.Fatalf("NewPersistentSagaCoordinator() error = %v", err)
	}
	defer coordinator.Close()

	saga := coordinator.Begin("Agent-E", "unsupported compensation")
	_, _ = coordinator.RegisterStep(saga.ID, "external", "call external-api", "called", "undo external-api")
	failed, err := coordinator.Abort(saga.ID, "test abort")
	if err == nil {
		t.Fatalf("Abort() error = nil, want unsupported compensation error")
	}
	if failed.Status != SagaStatusCompensationFailed {
		t.Fatalf("Abort() status = %s, want %s", failed.Status, SagaStatusCompensationFailed)
	}
}
