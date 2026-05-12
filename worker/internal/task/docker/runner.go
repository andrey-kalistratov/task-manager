package docker

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"os/exec"
	"time"

	"github.com/containerd/errdefs"
	"github.com/moby/moby/api/pkg/stdcopy"
	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/client"

	"github.com/andrey-kalistratov/task-manager/worker/internal/task"
)

var _ task.Executor = (*Runner)(nil)

type Runner struct {
	cli    *client.Client
	logger *slog.Logger
}

func NewRunner(logger *slog.Logger) (*Runner, error) {
	cmd := exec.Command("dockerd")
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("start dockerd: %w", err)
	}

	cli, err := client.New(client.FromEnv)
	if err != nil {
		return nil, fmt.Errorf("create docker client: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	for {
		if _, err = cli.Ping(ctx, client.PingOptions{}); err == nil {
			break
		}
		if ctx.Err() != nil {
			return nil, fmt.Errorf("dockerd did not start in time: %w", ctx.Err())
		}
		time.Sleep(100 * time.Millisecond)
	}

	return &Runner{
		cli:    cli,
		logger: logger,
	}, nil
}

func (r *Runner) Execute(ctx context.Context, spec task.ExecSpec) (task.Execution, error) {
	if err := r.ensureImage(ctx, spec.Image); err != nil {
		return task.Execution{}, fmt.Errorf("ensure image: %w", err)
	}

	resp, err := r.cli.ContainerCreate(ctx, client.ContainerCreateOptions{
		Config: &container.Config{
			Image:      spec.Image,
			Cmd:        []string{"bash", "-c", spec.Command},
			WorkingDir: spec.WorkDir,
		},
		HostConfig: &container.HostConfig{
			Binds:      []string{fmt.Sprintf("%s:%s", spec.WorkDir, spec.WorkDir)},
			AutoRemove: true,
		},
	})
	if err != nil {
		return task.Execution{}, fmt.Errorf("create container: %w", err)
	}

	if _, err = r.cli.ContainerStart(ctx, resp.ID, client.ContainerStartOptions{}); err != nil {
		return task.Execution{}, fmt.Errorf("start container: %w", err)
	}

	logs, err := r.cli.ContainerLogs(ctx, resp.ID, client.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     true,
	})
	if err != nil {
		return task.Execution{}, fmt.Errorf("get container logs: %w", err)
	}
	defer logs.Close()

	var stdout, stderr bytes.Buffer
	if _, err = stdcopy.StdCopy(&stdout, &stderr, logs); err != nil {
		return task.Execution{}, fmt.Errorf("parse container logs: %w", err)
	}

	result := r.cli.ContainerWait(ctx, resp.ID, client.ContainerWaitOptions{})
	select {
	case err = <-result.Error:
		return task.Execution{}, fmt.Errorf("wait container: %w", err)
	case resp := <-result.Result:
		r.logger.Info(
			"container finished", "exit_code", resp.StatusCode,
			"stdout", stdout.String(), "stderr", stderr.String(),
		)
		return task.Execution{ExitCode: int(resp.StatusCode)}, nil
	}
}

func (r *Runner) ensureImage(ctx context.Context, image string) error {
	_, err := r.cli.ImageInspect(ctx, image)
	switch {
	case err == nil:
		return nil
	case !errdefs.IsNotFound(err):
		return fmt.Errorf("inspect image: %w", err)
	}

	resp, err := r.cli.ImagePull(ctx, image, client.ImagePullOptions{})
	if err != nil {
		return fmt.Errorf("start image pull: %w", err)
	}
	defer func() {
		if err = resp.Close(); err != nil {
			r.logger.Error("failed to close docker response", "error", err)
		}
	}()
	if _, err = io.Copy(io.Discard, resp); err != nil {
		return fmt.Errorf("pull image: %w", err)
	}
	return nil
}

func (r *Runner) Close() error {
	return r.cli.Close()
}
