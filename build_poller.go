package goje

import (
	"context"
	"fmt"
	"time"
)

type InputHandler func(ctx context.Context, jobPath string, buildID int, actions []PendingInputAction) error

type BuildPoller struct {
	buildID         int
	inputHandler    InputHandler
	jenkins         *Jenkins
	jobPath         string
	pollingInterval time.Duration
	timeout         time.Duration
}

func newBuildPoller(jenkins *Jenkins, jobPath string, buildID int) *BuildPoller {
	return &BuildPoller{
		buildID:         buildID,
		jenkins:         jenkins,
		jobPath:         jobPath,
		pollingInterval: 5 * time.Second,
		timeout:         0,
	}
}

func (b *BuildPoller) WithPollingInterval(pollingInterval time.Duration) *BuildPoller {
	b.pollingInterval = pollingInterval
	return b
}

func (b *BuildPoller) WithTimeout(timeout time.Duration) *BuildPoller {
	b.timeout = timeout
	return b
}

func (b *BuildPoller) OnInput(handler InputHandler) *BuildPoller {
	b.inputHandler = handler
	return b
}

func (b *BuildPoller) Poll(ctx context.Context) (Build, error) {
	if b.timeout != 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, b.timeout)
		defer cancel()
	}

	for {
		build, err := b.jenkins.GetBuild(ctx, b.jobPath, b.buildID)
		if err != nil {
			return Build{}, fmt.Errorf("failed to get build: %w", err)
		}

		if !build.Building {
			return build, nil
		}

		if b.inputHandler != nil {
			pendingInputs, err := b.jenkins.GetPendingInputActions(ctx, b.jobPath, b.buildID)
			if err != nil {
				return Build{}, fmt.Errorf("failed to check if build is paused for input: %w", err)
			}

			if len(pendingInputs) > 0 {
				if err := b.inputHandler(ctx, b.jobPath, b.buildID, pendingInputs); err != nil {
					return Build{}, fmt.Errorf("failed to invoke input handler: %w", err)
				}
			}
		}

		select {
		case <-time.After(b.pollingInterval):
		case <-ctx.Done():
			return Build{}, ctx.Err()
		}
	}
}
