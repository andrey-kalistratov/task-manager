package daemon

import (
	"database/sql"
	"fmt"
	"log/slog"

	"github.com/andrey-kalistratov/task-manager/planner/filelogger"
	"github.com/andrey-kalistratov/task-manager/planner/internal/config"
	"github.com/andrey-kalistratov/task-manager/planner/internal/task"
	"github.com/andrey-kalistratov/task-manager/planner/internal/task/ipc"
	"github.com/andrey-kalistratov/task-manager/planner/internal/task/sqlite"
	"github.com/andrey-kalistratov/task-manager/planner/unixsocket"

	_ "github.com/mattn/go-sqlite3"
)

func Run(cfg *config.Config) error {
	logger, cleanup, err := filelogger.New(cfg.Logging.File, &slog.HandlerOptions{
		Level: cfg.Logging.Level,
	})
	if err != nil {
		return fmt.Errorf("failed to setup logger: %w", err)
	}
	defer cleanup()

	db, err := sql.Open("sqlite3", cfg.DB.File)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %v", err)
	}
	defer db.Close()

	storage := sqlite.NewStorage(db)
	service := task.NewService(storage)

	router := unixsocket.NewRouter()
	router.Register("add", ipc.NewAddHandler(service))

	server := unixsocket.NewServer(config.UnixSocket, router, logger)
	return server.ListenAndServe()
}
