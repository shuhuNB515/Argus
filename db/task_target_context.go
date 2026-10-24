package db

import (
	"database/sql"
	"encoding/json"
)

// TaskTargetContext is the persisted, executor-facing authorization contract.
// It intentionally contains no credentials or resolved IPs.
type TaskTargetContext struct {
	Version         int      `json:"version"`
	Mode            string   `json:"mode"`
	Target          string   `json:"target"`
	TestRange       []string `json:"test_range"`
	Ports           []int    `json:"ports"`
	AllowSubdomains bool     `json:"allow_subdomains"`
}

func (d *DB) UpsertTaskTargetContext(taskID int64, ctx TaskTargetContext) error {
	raw, err := json.Marshal(ctx)
	if err != nil {
		return err
	}
	_, err = d.Exec(`
INSERT INTO task_target_context(task_id, context)
VALUES ($1,$2::jsonb)
ON CONFLICT (task_id) DO UPDATE SET context=EXCLUDED.context, updated_at=now()`, taskID, string(raw))
	return err
}

func (d *DB) TaskTargetContext(taskID int64) (*TaskTargetContext, error) {
	var raw []byte
	err := d.QueryRow(`SELECT context FROM task_target_context WHERE task_id=$1`, taskID).Scan(&raw)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var ctx TaskTargetContext
	if err := json.Unmarshal(raw, &ctx); err != nil {
		return nil, err
	}
	return &ctx, nil
}
