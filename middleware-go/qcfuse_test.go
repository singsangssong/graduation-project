package main

import (
	"testing"

	pb "agenic-middleware/middleware-go/pb"
)

func TestGroupReadTasksByResourceDoesNotFuseDifferentResources(t *testing.T) {
	tasks := []ReadTask{
		{Req: &pb.ReadRequest{ResourceId: "ticket-A", Intent: "purchase"}},
		{Req: &pb.ReadRequest{ResourceId: "ticket-A", Intent: "purchase"}},
		{Req: &pb.ReadRequest{ResourceId: "ticket-B", Intent: "purchase"}},
	}

	groups := groupReadTasksByResource(tasks)

	if len(groups) != 2 {
		t.Fatalf("group count = %d, want 2", len(groups))
	}
	if len(groups[readFusionKey(tasks[0].Req)]) != 2 {
		t.Fatalf("ticket-A group size = %d, want 2", len(groups[readFusionKey(tasks[0].Req)]))
	}
}

func TestGroupReadTasksForModeKeepsBaselineRequestsSeparate(t *testing.T) {
	tasks := []ReadTask{
		{Req: &pb.ReadRequest{AgentId: "A", ResourceId: "ticket-A", Intent: "purchase"}},
		{Req: &pb.ReadRequest{AgentId: "B", ResourceId: "ticket-A", Intent: "purchase"}},
	}

	baseline := groupReadTasksForMode(tasks, ExperimentModeBaseline)
	qcfuse := groupReadTasksForMode(tasks, ExperimentModeQCFuse)

	if len(baseline) != 2 {
		t.Fatalf("baseline group count = %d, want 2", len(baseline))
	}
	if len(qcfuse) != 1 {
		t.Fatalf("qcfuse group count = %d, want 1", len(qcfuse))
	}
}
