package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/dakotalillie/goje"
)

func buildJobWithParameters(jenkins *goje.Jenkins) error {
	var (
		ctx        = context.Background()
		jobPath    = "my-folder/test3"
		parameters = map[string]string{"foo": "bar"}
	)

	slog.Info("building job", "job", jobPath)

	queueItemID, err := jenkins.BuildJobWithParameters(ctx, jobPath, parameters)
	if err != nil {
		return fmt.Errorf("failed to build job: %w", err)
	}

	slog.Info("build enqueued")

	buildID, err := jenkins.NewQueueItemPoller(queueItemID).Poll(ctx)
	if err != nil {
		return fmt.Errorf("failed to poll queue item: %w", err)
	}

	build, err := jenkins.GetBuild(ctx, jobPath, buildID)
	if err != nil {
		return fmt.Errorf("failed to get build: %w", err)
	}

	slog.Info("build started", "url", build.URL)

	printLogs := func(logs string) error {
		fmt.Println(logs)
		return nil
	}

	build, err = jenkins.NewBuildPoller(jobPath, buildID).OnLogs(printLogs).Poll(ctx)
	if err != nil {
		return fmt.Errorf("failed to poll build: %w", err)
	}

	if build.Result != goje.BuildResultSuccess {
		return fmt.Errorf("build did not succeed: got %s", build.Result)
	}

	slog.Info("build succeeded")
	return nil
}

func buildJobWithInput(jenkins *goje.Jenkins) error {
	var (
		ctx     = context.Background()
		jobPath = "my-folder/foo"
	)

	slog.Info("building job", "job", jobPath)

	queueItemID, err := jenkins.BuildJob(ctx, jobPath)
	if err != nil {
		return fmt.Errorf("failed to build job: %w", err)
	}

	slog.Info("build enqueued")

	buildID, err := jenkins.NewQueueItemPoller(queueItemID).Poll(ctx)
	if err != nil {
		return fmt.Errorf("failed to poll queue item: %w", err)
	}

	build, err := jenkins.GetBuild(ctx, jobPath, buildID)
	if err != nil {
		return fmt.Errorf("failed to get build: %w", err)
	}

	slog.Info("build started", "url", build.URL)

	build, err = jenkins.NewBuildPoller(jobPath, buildID).OnInput(jenkins.Proceed()).Poll(ctx)
	if err != nil {
		return fmt.Errorf("failed to poll build: %w", err)
	}

	if build.Result != goje.BuildResultSuccess {
		return fmt.Errorf("build did not succeed: got %s", build.Result)
	}

	slog.Info("build succeeded")
	return nil
}

func main() {
	username, password, ok := strings.Cut(os.Getenv("JENKINS_CREDENTIALS"), ":")
	if !ok {
		slog.Error("missing JENKINS_CREDENTIALS environment variable")
		return
	}

	jenkins := goje.NewJenkins("http://localhost:8080").WithBasicAuth(username, password)

	if err := buildJobWithParameters(jenkins); err != nil {
		slog.Error("failed to build job with parameters", "error", err)
		os.Exit(1)
	}

	if err := buildJobWithInput(jenkins); err != nil {
		slog.Error("failed to build job with input", "error", err)
		os.Exit(1)
	}
}
