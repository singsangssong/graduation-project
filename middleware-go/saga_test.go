package main

import "testing"

func TestSagaCoordinatorLifecycle(t *testing.T) {
	coordinator := NewSagaCoordinator()

	saga := coordinator.Begin("Agent-A", "purchase flight ticket")
	if saga.Status != SagaStatusActive {
		t.Fatalf("Begin() status = %s, want %s", saga.Status, SagaStatusActive)
	}

	registered, err := coordinator.RegisterStep(
		saga.ID,
		"reserve-ticket",
		"reserve ticket",
		"ticket reserved",
		"release ticket",
	)
	if err != nil {
		t.Fatalf("RegisterStep() error = %v", err)
	}
	if len(registered.Steps) != 1 || registered.Steps[0].Status != SagaStepStatusCompleted {
		t.Fatalf("RegisterStep() saga = %+v", registered)
	}

	validated, err := coordinator.Validate(saga.ID)
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
	if validated.Status != SagaStatusValidated {
		t.Fatalf("Validate() status = %s, want %s", validated.Status, SagaStatusValidated)
	}

	committed, err := coordinator.Commit(saga.ID)
	if err != nil {
		t.Fatalf("Commit() error = %v", err)
	}
	if committed.Status != SagaStatusCommitted {
		t.Fatalf("Commit() status = %s, want %s", committed.Status, SagaStatusCommitted)
	}
}

func TestSagaCoordinatorAbortCompensatesCompletedStepsInReverseOrder(t *testing.T) {
	coordinator := NewSagaCoordinator()
	saga := coordinator.Begin("Agent-B", "purchase flight ticket")

	_, _ = coordinator.RegisterStep(saga.ID, "reserve", "reserve ticket", "reserved", "release ticket")
	_, _ = coordinator.RegisterStep(saga.ID, "charge", "charge card", "charged", "refund card")

	aborted, err := coordinator.Abort(saga.ID, "ATCC conflict")
	if err != nil {
		t.Fatalf("Abort() error = %v", err)
	}
	if aborted.Status != SagaStatusCompensated {
		t.Fatalf("Abort() status = %s, want %s", aborted.Status, SagaStatusCompensated)
	}
	if aborted.Steps[0].Status != SagaStepStatusCompensated || aborted.Steps[1].Status != SagaStepStatusCompensated {
		t.Fatalf("Abort() did not compensate steps: %+v", aborted.Steps)
	}
	if aborted.CompensationLog[0] != "refund card" || aborted.CompensationLog[1] != "release ticket" {
		t.Fatalf("compensation order = %v, want [refund card release ticket]", aborted.CompensationLog)
	}
}

func TestSagaCoordinatorValidationRejectsEmptySaga(t *testing.T) {
	coordinator := NewSagaCoordinator()
	saga := coordinator.Begin("Agent-C", "empty workflow")

	validated, err := coordinator.Validate(saga.ID)
	if err == nil {
		t.Fatalf("Validate() error = nil, want validation error")
	}
	if validated.Status != SagaStatusValidationFailed {
		t.Fatalf("Validate() status = %s, want %s", validated.Status, SagaStatusValidationFailed)
	}
}

func TestSagaCoordinatorCanRecoverAfterValidationFailure(t *testing.T) {
	coordinator := NewSagaCoordinator()
	saga := coordinator.Begin("Agent-D", "recoverable workflow")

	_, _ = coordinator.Validate(saga.ID)
	registered, err := coordinator.RegisterStep(
		saga.ID,
		"fixed-step",
		"repair missing action",
		"repaired",
		"undo repair",
	)
	if err != nil {
		t.Fatalf("RegisterStep() after validation failure error = %v", err)
	}
	if registered.Status != SagaStatusActive {
		t.Fatalf("RegisterStep() status = %s, want %s", registered.Status, SagaStatusActive)
	}

	validated, err := coordinator.Validate(saga.ID)
	if err != nil {
		t.Fatalf("Validate() after repair error = %v", err)
	}
	if validated.Status != SagaStatusValidated {
		t.Fatalf("Validate() status = %s, want %s", validated.Status, SagaStatusValidated)
	}
}
