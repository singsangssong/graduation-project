package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

type SagaEvent struct {
	ID        int64                  `json:"id"`
	SagaID    string                 `json:"saga_id"`
	Type      string                 `json:"type"`
	Metadata  map[string]interface{} `json:"metadata"`
	CreatedAt time.Time              `json:"created_at"`
}

type SQLiteStore struct {
	db           *sql.DB
	initialStock int
}

func NewSQLiteStore(path string, initialStock int) (*SQLiteStore, error) {
	if path != ":memory:" {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return nil, err
		}
	}
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	store := &SQLiteStore{db: db, initialStock: initialStock}
	if err := store.migrate(); err != nil {
		_ = db.Close()
		return nil, err
	}
	return store, nil
}

func (s *SQLiteStore) migrate() error {
	_, err := s.db.Exec(`
		PRAGMA journal_mode = WAL;
		PRAGMA busy_timeout = 5000;
		CREATE TABLE IF NOT EXISTS sagas (
			saga_id TEXT PRIMARY KEY,
			agent_id TEXT NOT NULL,
			goal TEXT NOT NULL,
			status TEXT NOT NULL,
			validation_message TEXT NOT NULL,
			abort_reason TEXT NOT NULL,
			created_at INTEGER NOT NULL,
			updated_at INTEGER NOT NULL
		);
		CREATE TABLE IF NOT EXISTS saga_steps (
			saga_id TEXT NOT NULL,
			sequence INTEGER NOT NULL,
			step_id TEXT NOT NULL,
			action TEXT NOT NULL,
			result TEXT NOT NULL,
			compensation_action TEXT NOT NULL,
			status TEXT NOT NULL,
			PRIMARY KEY (saga_id, step_id)
		);
		CREATE TABLE IF NOT EXISTS saga_events (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			saga_id TEXT NOT NULL,
			event_type TEXT NOT NULL,
			metadata_json TEXT NOT NULL,
			created_at INTEGER NOT NULL
		);
		CREATE TABLE IF NOT EXISTS resources (
			resource_id TEXT PRIMARY KEY,
			available_stock INTEGER NOT NULL
		);
		CREATE TABLE IF NOT EXISTS resource_reservations (
			saga_id TEXT NOT NULL,
			resource_id TEXT NOT NULL,
			status TEXT NOT NULL,
			PRIMARY KEY (saga_id, resource_id)
		);
	`)
	return err
}

func (s *SQLiteStore) SaveSaga(saga SagaRecord) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec(`
		INSERT INTO sagas (saga_id, agent_id, goal, status, validation_message, abort_reason, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(saga_id) DO UPDATE SET
			agent_id=excluded.agent_id, goal=excluded.goal, status=excluded.status,
			validation_message=excluded.validation_message, abort_reason=excluded.abort_reason,
			updated_at=excluded.updated_at
	`, saga.ID, saga.AgentID, saga.Goal, saga.Status, saga.ValidationMessage, saga.AbortReason, saga.CreatedAt.UnixNano(), saga.UpdatedAt.UnixNano())
	if err != nil {
		return err
	}
	if _, err = tx.Exec("DELETE FROM saga_steps WHERE saga_id = ?", saga.ID); err != nil {
		return err
	}
	for sequence, step := range saga.Steps {
		_, err = tx.Exec(`
			INSERT INTO saga_steps (saga_id, sequence, step_id, action, result, compensation_action, status)
			VALUES (?, ?, ?, ?, ?, ?, ?)
		`, saga.ID, sequence, step.ID, step.Action, step.Result, step.CompensationAction, step.Status)
		if err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (s *SQLiteStore) LoadSagas() ([]SagaRecord, error) {
	rows, err := s.db.Query(`
		SELECT saga_id, agent_id, goal, status, validation_message, abort_reason, created_at, updated_at
		FROM sagas ORDER BY created_at
	`)
	if err != nil {
		return nil, err
	}
	var sagas []SagaRecord
	for rows.Next() {
		var saga SagaRecord
		var createdAt, updatedAt int64
		if err := rows.Scan(&saga.ID, &saga.AgentID, &saga.Goal, &saga.Status, &saga.ValidationMessage, &saga.AbortReason, &createdAt, &updatedAt); err != nil {
			return nil, err
		}
		saga.CreatedAt = time.Unix(0, createdAt)
		saga.UpdatedAt = time.Unix(0, updatedAt)
		sagas = append(sagas, saga)
	}
	if err := rows.Err(); err != nil {
		_ = rows.Close()
		return nil, err
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	for i := range sagas {
		steps, err := s.loadSteps(sagas[i].ID)
		if err != nil {
			return nil, err
		}
		sagas[i].Steps = steps
	}
	return sagas, nil
}

func (s *SQLiteStore) loadSteps(sagaID string) ([]SagaStepRecord, error) {
	rows, err := s.db.Query(`
		SELECT step_id, action, result, compensation_action, status
		FROM saga_steps WHERE saga_id = ? ORDER BY sequence
	`, sagaID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var steps []SagaStepRecord
	for rows.Next() {
		var step SagaStepRecord
		if err := rows.Scan(&step.ID, &step.Action, &step.Result, &step.CompensationAction, &step.Status); err != nil {
			return nil, err
		}
		steps = append(steps, step)
	}
	return steps, rows.Err()
}

func (s *SQLiteStore) AppendEvent(sagaID, eventType string, metadata map[string]interface{}) error {
	raw, err := json.Marshal(metadata)
	if err != nil {
		return err
	}
	_, err = s.db.Exec(
		"INSERT INTO saga_events (saga_id, event_type, metadata_json, created_at) VALUES (?, ?, ?, ?)",
		sagaID, eventType, string(raw), time.Now().UnixNano(),
	)
	return err
}

func (s *SQLiteStore) Events(sagaID string) ([]SagaEvent, error) {
	rows, err := s.db.Query(`
		SELECT id, saga_id, event_type, metadata_json, created_at
		FROM saga_events WHERE saga_id = ? ORDER BY id
	`, sagaID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []SagaEvent
	for rows.Next() {
		var event SagaEvent
		var raw string
		var createdAt int64
		if err := rows.Scan(&event.ID, &event.SagaID, &event.Type, &raw, &createdAt); err != nil {
			return nil, err
		}
		event.Metadata = make(map[string]interface{})
		_ = json.Unmarshal([]byte(raw), &event.Metadata)
		event.CreatedAt = time.Unix(0, createdAt)
		events = append(events, event)
	}
	return events, rows.Err()
}

func (s *SQLiteStore) ExecuteAction(sagaID, action string) error {
	verb, resourceID := parseResourceAction(action)
	if verb != "reserve" {
		return nil
	}
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if err := ensureResource(tx, resourceID, s.initialStock); err != nil {
		return err
	}
	var status string
	err = tx.QueryRow("SELECT status FROM resource_reservations WHERE saga_id = ? AND resource_id = ?", sagaID, resourceID).Scan(&status)
	if err == nil {
		return tx.Commit()
	}
	if err != sql.ErrNoRows {
		return err
	}
	result, err := tx.Exec("UPDATE resources SET available_stock = available_stock - 1 WHERE resource_id = ? AND available_stock > 0", resourceID)
	if err != nil {
		return err
	}
	affected, _ := result.RowsAffected()
	if affected == 0 {
		return fmt.Errorf("resource %s has no available stock", resourceID)
	}
	if _, err := tx.Exec("INSERT INTO resource_reservations (saga_id, resource_id, status) VALUES (?, ?, 'RESERVED')", sagaID, resourceID); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *SQLiteStore) ExecuteCompensation(sagaID, action string) error {
	verb, resourceID := parseResourceAction(action)
	if verb == "" {
		return fmt.Errorf("unsupported compensation action %q", action)
	}
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	result, err := tx.Exec("UPDATE resource_reservations SET status = 'RELEASED' WHERE saga_id = ? AND resource_id = ? AND status = 'RESERVED'", sagaID, resourceID)
	if err != nil {
		return err
	}
	affected, _ := result.RowsAffected()
	if affected > 0 {
		if _, err := tx.Exec("UPDATE resources SET available_stock = available_stock + 1 WHERE resource_id = ?", resourceID); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (s *SQLiteStore) CommitReservations(sagaID string) error {
	_, err := s.db.Exec("UPDATE resource_reservations SET status = 'COMMITTED' WHERE saga_id = ? AND status = 'RESERVED'", sagaID)
	return err
}

func (s *SQLiteStore) ResourceStock(resourceID string) (int, error) {
	if _, err := s.db.Exec("INSERT OR IGNORE INTO resources (resource_id, available_stock) VALUES (?, ?)", resourceID, s.initialStock); err != nil {
		return 0, err
	}
	var stock int
	err := s.db.QueryRow("SELECT available_stock FROM resources WHERE resource_id = ?", resourceID).Scan(&stock)
	return stock, err
}

func (s *SQLiteStore) Reset() error {
	_, err := s.db.Exec(`
		DELETE FROM resource_reservations;
		DELETE FROM resources;
		DELETE FROM saga_events;
		DELETE FROM saga_steps;
		DELETE FROM sagas;
	`)
	return err
}

func (s *SQLiteStore) Close() error {
	return s.db.Close()
}

func ensureResource(tx *sql.Tx, resourceID string, initialStock int) error {
	_, err := tx.Exec("INSERT OR IGNORE INTO resources (resource_id, available_stock) VALUES (?, ?)", resourceID, initialStock)
	return err
}

func parseResourceAction(action string) (string, string) {
	parts := strings.Fields(strings.TrimSpace(action))
	if len(parts) != 2 {
		return "", ""
	}
	verb := strings.ToLower(parts[0])
	if verb != "reserve" && verb != "release" {
		return "", ""
	}
	return verb, parts[1]
}
