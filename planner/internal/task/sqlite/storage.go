package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"

	"github.com/andrey-kalistratov/task-manager/planner/internal/config"
	"github.com/andrey-kalistratov/task-manager/planner/internal/task"
)

var _ task.Storage = (*Storage)(nil)

type Storage struct {
	db     *sql.DB
	logger *slog.Logger
}

func NewStorage(cfg *config.Config, logger *slog.Logger) (s *Storage, err error) {
	s = &Storage{logger: logger}

	defer func() {
		if err != nil {
			s.releaseResources()
		}
	}()

	const opts = "?_loc=UTC"
	s.db, err = sql.Open("sqlite3", cfg.Storage.SqliteFile+opts)
	if err != nil {
		return s, fmt.Errorf("open sqlite db: %w", err)
	}

	if err = initSchema(s.db); err != nil {
		return s, fmt.Errorf("init sql schema: %w", err)
	}

	return
}

func initSchema(db *sql.DB) error {
	const schema = `CREATE TABLE IF NOT EXISTS tasks
(
    id         TEXT PRIMARY KEY,
    status     TEXT NOT NULL,
    created_at DATETIME NOT NULL,
    command    TEXT NOT NULL,
    name       TEXT NOT NULL,
    image      TEXT NOT NULL,
    inputs     TEXT NOT NULL DEFAULT '{}',
    uploads    TEXT NOT NULL DEFAULT '{}',
    downloads  TEXT NOT NULL DEFAULT '{}',
    outputs    TEXT NOT NULL DEFAULT '{}'
);`

	_, err := db.Exec(schema)
	return err
}

func (s *Storage) Save(ctx context.Context, t *task.Task) error {
	const query = `INSERT INTO tasks
    (id, status, created_at, command, name, image, inputs, uploads, downloads, outputs)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
ON CONFLICT(id) DO UPDATE SET
    status = excluded.status,
    created_at = excluded.created_at,
    command = excluded.command,
    name = excluded.name,
    image = excluded.image,
    inputs = excluded.inputs,
    uploads = excluded.uploads,
    downloads = excluded.downloads,
    outputs = excluded.outputs`

	dto, err := newTask(t)
	if err != nil {
		return fmt.Errorf("serialize task: %w", err)
	}

	_, err = s.db.ExecContext(
		ctx,
		query,
		dto.ID,
		dto.Status,
		dto.CreatedAt,
		dto.Command,
		dto.Name,
		dto.Image,
		dto.Inputs,
		dto.Uploads,
		dto.Downloads,
		dto.Outputs,
	)
	if err != nil {
		return fmt.Errorf("insert task: %w", err)
	}
	return nil
}

func (s *Storage) Get(ctx context.Context, id uuid.UUID) (*task.Task, error) {
	const query = `SELECT id, status, created_at, command, name, image, inputs, uploads, downloads, outputs 
FROM tasks 
WHERE id = ?`

	var dto Task
	row := s.db.QueryRowContext(ctx, query, id.String())
	err := row.Scan(
		&dto.ID,
		&dto.Status,
		&dto.CreatedAt,
		&dto.Command,
		&dto.Name,
		&dto.Image,
		&dto.Inputs,
		&dto.Uploads,
		&dto.Downloads,
		&dto.Outputs,
	)
	if err != nil {
		return nil, fmt.Errorf("read row: %w", err)
	}

	t, err := dto.toModel()
	if err != nil {
		return nil, fmt.Errorf("deserialize task: %w", err)
	}
	return t, nil
}

func (s *Storage) releaseResources() {
	if err := s.db.Close(); err != nil {
		s.logger.Error("failed to close db", "error", err)
	}
}

func (s *Storage) Close() error {
	if err := s.db.Close(); err != nil {
		return fmt.Errorf("close sqlite db: %w", err)
	}
	return nil
}
