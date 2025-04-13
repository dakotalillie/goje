package main

import (
	"context"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/dakotalillie/goje"
)

func main() {
	username, password, ok := strings.Cut(os.Getenv("JENKINS_CREDENTIALS"), ":")
	if !ok {
		slog.Error("missing JENKINS_CREDENTIALS environment variable")
		return
	}

	var (
		ctx     = context.Background()
		jenkins = goje.New("http://localhost:8080").WithBasicAuth(username, password)
		jobPath = "my-folder/foo"
	)

	queueItemID, err := jenkins.BuildJob(ctx, jobPath)
	if err != nil {
		slog.Error("failed to build job", "error", err)
		return
	}

	buildID, err := jenkins.NewQueueItemPoller(queueItemID).WithTimeout(5 * time.Minute).Poll(ctx)
	if err != nil {
		slog.Error("failed to poll queue item", "error", err)
		return
	}

	build, err := jenkins.NewBuildPoller(jobPath, buildID).OnInput(jenkins.Proceed()).Poll(ctx)
	if err != nil {
		slog.Error("failed to poll build", "error", err)
		return
	}

	if build.Result != goje.BuildResultSuccess {
		slog.Error("build did not succeed", "result", build.Result)
		return
	}

	slog.Info("build succeeded", "job", jobPath, "build_id", buildID, "result", build.Result)
}
