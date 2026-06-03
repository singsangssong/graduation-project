package main

import (
	"math"
	"testing"

	pb "agenic-middleware/middleware-go/pb"
)

func testConfig() Config {
	return Config{
		TokenCostWeight:   0.002,
		LatencyCostWeight: 0.5,
	}
}

func TestCalculateSunkCostWithConfig(t *testing.T) {
	req := &pb.CommitRequest{
		AgentId:             "Agent-A",
		AccumulatedTokens:   2100,
		InferenceLatencySec: 10,
	}

	got := calculateSunkCostWithConfig(req, testConfig())
	want := float32(9.2)

	if math.Abs(float64(got-want)) > 0.0001 {
		t.Fatalf("calculateSunkCostWithConfig() = %.2f, want %.2f", got, want)
	}
}

func TestRankCommitTasksOrdersHighestCostFirst(t *testing.T) {
	batch := []CommitTask{
		{Req: &pb.CommitRequest{AgentId: "Agent-low", AccumulatedTokens: 200, InferenceLatencySec: 1.2}},
		{Req: &pb.CommitRequest{AgentId: "Agent-high", AccumulatedTokens: 4500, InferenceLatencySec: 8.5}},
		{Req: &pb.CommitRequest{AgentId: "Agent-mid", AccumulatedTokens: 1800, InferenceLatencySec: 3.0}},
	}

	ranked := rankCommitTasks(batch, testConfig())

	if ranked[0].Req.GetAgentId() != "Agent-high" {
		t.Fatalf("winner = %s, want Agent-high", ranked[0].Req.GetAgentId())
	}
	if batch[0].Req.GetAgentId() != "Agent-low" {
		t.Fatalf("rankCommitTasks mutated input batch")
	}
}

func TestMetricsRecordFusedReadBatchCountsSavedDBReads(t *testing.T) {
	m := NewMiddlewareMetrics()

	m.RecordFusedReadBatch(10)
	snapshot := m.Snapshot()

	if snapshot.FusedReadBatches != 1 {
		t.Fatalf("FusedReadBatches = %d, want 1", snapshot.FusedReadBatches)
	}
	if snapshot.SavedDBReads != 9 {
		t.Fatalf("SavedDBReads = %d, want 9", snapshot.SavedDBReads)
	}
}

func TestMetricsResetClearsCounters(t *testing.T) {
	m := NewMiddlewareMetrics()

	m.RecordReadRequest()
	m.RecordFusedReadBatch(5)
	m.RecordApprovedCommit("Agent-A", 13.25)
	m.RecordRolledBackCommit(2.5)
	m.Reset()
	snapshot := m.Snapshot()

	if snapshot.ReadRequests != 0 || snapshot.SavedDBReads != 0 || snapshot.ApprovedCommits != 0 || snapshot.TotalSavedCostUsd != 0 {
		t.Fatalf("Reset() did not clear metrics: %+v", snapshot)
	}
}
