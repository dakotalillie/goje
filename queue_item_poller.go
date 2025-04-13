package goje

import (
	"context"
	"errors"
	"fmt"
	"time"
)

type QueueItemPoller struct {
	jenkins         *Jenkins
	pollingInterval time.Duration
	queueItemID     int
	timeout         time.Duration
}

func newQueueItemPoller(jenkins *Jenkins, queueItemID int) *QueueItemPoller {
	return &QueueItemPoller{
		jenkins:         jenkins,
		pollingInterval: time.Second,
		queueItemID:     queueItemID,
		timeout:         0,
	}
}

func (q *QueueItemPoller) WithPollingInterval(pollingInterval time.Duration) *QueueItemPoller {
	q.pollingInterval = pollingInterval
	return q
}

func (q *QueueItemPoller) WithTimeout(timeout time.Duration) *QueueItemPoller {
	q.timeout = timeout
	return q
}

func (q *QueueItemPoller) Poll(ctx context.Context) (int, error) {
	if q.timeout != 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, q.timeout)
		defer cancel()
	}

	for {
		queueItem, err := q.jenkins.GetQueueItem(ctx, q.queueItemID)
		if err != nil {
			return 0, fmt.Errorf("failed to get queue item: %w", err)
		}

		if queueItem.Cancelled {
			return 0, errors.New("job was cancelled")
		}

		if queueItem.Executable.Number != 0 {
			return queueItem.Executable.Number, nil
		}

		select {
		case <-time.After(q.pollingInterval):
		case <-ctx.Done():
			return 0, ctx.Err()
		}
	}
}
