package daemon

import (
	"database/sql"
	"fmt"
	"log/slog"

	"task-manager/planner/filelogger"
	"task-manager/planner/internal/config"
	"task-manager/planner/internal/task/ipc"
	"task-manager/planner/internal/task/sqlite"
	"task-manager/planner/unixsocket"

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

	router := unixsocket.NewRouter()
	router.Register("add", ipc.NewAddHandler(storage))

	server := unixsocket.NewServer(config.UnixSocket, router, logger)
	return server.ListenAndServe()
}
