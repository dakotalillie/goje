package goje

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"
)

// Jenkins is a client for the Jenkins API
type Jenkins struct {
	baseURL  string
	client   *http.Client
	withAuth func(req *http.Request)
}

// New creates a new Jenkins client
func New(baseURL string) *Jenkins {
	return &Jenkins{
		baseURL:  baseURL,
		client:   http.DefaultClient,
		withAuth: nil,
	}
}

// WithBasicAuth sets the basic auth for the Jenkins client
func (j *Jenkins) WithBasicAuth(username, password string) *Jenkins {
	j.withAuth = func(req *http.Request) {
		req.SetBasicAuth(username, password)
	}
	return j
}

// WithHTTPClient sets the HTTP client for the Jenkins client
func (j *Jenkins) WithHTTPClient(client *http.Client) *Jenkins {
	j.client = client
	return j
}

// BuildJob triggers a build for the specified job without parameters and returns the queue item ID
func (j *Jenkins) BuildJob(ctx context.Context, jobPath string) (int, error) {
	u, err := url.Parse(j.baseURL)
	if err != nil {
		return 0, fmt.Errorf("failed to parse base url: %w", err)
	}

	u = u.JoinPath("/job/", j.parseJobPath(jobPath), "/build")

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u.String(), nil)
	if err != nil {
		return 0, fmt.Errorf("failed to create request: %w", err)
	}

	if j.withAuth != nil {
		j.withAuth(req)
	}

	res, err := j.client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("failed to perform request: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusCreated {
		return 0, fmt.Errorf("unexpected status code: %d", res.StatusCode)
	}

	queueItemID, err := strconv.Atoi(path.Base(path.Clean(res.Header.Get("Location"))))
	if err != nil {
		return 0, fmt.Errorf("failed to parse queue item ID: %w", err)
	}

	return queueItemID, nil
}

// ListJobs lists all jobs in the Jenkins instance
func (j *Jenkins) ListJobs(ctx context.Context) ([]Job, error) {
	u, err := url.Parse(j.baseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse base url: %w", err)
	}

	u = u.JoinPath("/api/json")

	q := u.Query()
	q.Set("tree", "jobs[name,url,color]")
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if j.withAuth != nil {
		j.withAuth(req)
	}

	res, err := j.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to perform request: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", res.StatusCode)
	}

	var jobsResponse struct {
		Jobs []Job `json:"jobs"`
	}
	if err := json.NewDecoder(res.Body).Decode(&jobsResponse); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return jobsResponse.Jobs, nil
}

// GetBuild gets a Jenkins build
func (j *Jenkins) GetBuild(ctx context.Context, jobPath string, buildID int) (Build, error) {
	u, err := url.Parse(j.baseURL)
	if err != nil {
		return Build{}, fmt.Errorf("failed to parse base url: %w", err)
	}

	u = u.JoinPath("/job/", j.parseJobPath(jobPath), strconv.Itoa(buildID), "/api/json")

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return Build{}, fmt.Errorf("failed to create request: %w", err)
	}

	if j.withAuth != nil {
		j.withAuth(req)
	}

	res, err := j.client.Do(req)
	if err != nil {
		return Build{}, fmt.Errorf("failed to perform request: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return Build{}, fmt.Errorf("unexpected status code: %d", res.StatusCode)
	}

	var build Build
	if err := json.NewDecoder(res.Body).Decode(&build); err != nil {
		return Build{}, fmt.Errorf("failed to decode response: %w", err)
	}

	return build, nil
}

// GetPendingInputActions gets the pending input actions for a Jenkins build
func (j *Jenkins) GetPendingInputActions(ctx context.Context, jobPath string, buildID int) ([]PendingInputAction, error) {
	u, err := url.Parse(j.baseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse base url: %w", err)
	}

	u = u.JoinPath("/job/", j.parseJobPath(jobPath), strconv.Itoa(buildID), "/wfapi/pendingInputActions")

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if j.withAuth != nil {
		j.withAuth(req)
	}

	res, err := j.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to perform request: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", res.StatusCode)
	}

	var pendingInputActions []PendingInputAction
	if err := json.NewDecoder(res.Body).Decode(&pendingInputActions); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return pendingInputActions, nil
}

func (j *Jenkins) GetQueueItem(ctx context.Context, queueID int) (QueueItem, error) {
	u, err := url.Parse(j.baseURL)
	if err != nil {
		return QueueItem{}, fmt.Errorf("failed to parse base url: %w", err)
	}

	u = u.JoinPath("/queue/item/", strconv.Itoa(queueID), "/api/json")

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return QueueItem{}, fmt.Errorf("failed to create request: %w", err)
	}

	if j.withAuth != nil {
		j.withAuth(req)
	}

	res, err := j.client.Do(req)
	if err != nil {
		return QueueItem{}, fmt.Errorf("failed to perform request: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return QueueItem{}, fmt.Errorf("unexpected status code: %d", res.StatusCode)
	}

	var queueItem QueueItem
	if err := json.NewDecoder(res.Body).Decode(&queueItem); err != nil {
		return QueueItem{}, fmt.Errorf("failed to decode response: %w", err)
	}

	return queueItem, nil
}

func (j *Jenkins) NewBuildPoller(jobPath string, buildID int) *BuildPoller {
	return newBuildPoller(j, jobPath, buildID)
}

func (j *Jenkins) NewQueueItemPoller(queueItemID int) *QueueItemPoller {
	return newQueueItemPoller(j, queueItemID)
}

func (j *Jenkins) Proceed() InputHandler {
	return func(ctx context.Context, jobPath string, buildID int, actions []PendingInputAction) error {
		u, err := url.Parse(j.baseURL)
		if err != nil {
			return fmt.Errorf("failed to parse base url: %w", err)
		}

		for _, action := range actions {
			u := u.JoinPath("/job/", j.parseJobPath(jobPath), strconv.Itoa(buildID), "/input/", action.ID, "/proceedEmpty")

			req, err := http.NewRequestWithContext(ctx, http.MethodPost, u.String(), nil)
			if err != nil {
				return fmt.Errorf("failed to create request: %w", err)
			}

			if j.withAuth != nil {
				j.withAuth(req)
			}

			res, err := j.client.Do(req)
			if err != nil {
				return fmt.Errorf("failed to perform request: %w", err)
			}
			defer res.Body.Close()

			if res.StatusCode != http.StatusOK {
				return fmt.Errorf("unexpected status code: %d", res.StatusCode)
			}
		}

		return nil
	}
}

func (j *Jenkins) parseJobPath(jobPath string) string {
	parts := strings.Split(strings.Trim(path.Clean(jobPath), "/"), "/")
	return strings.Join(parts, "/job/")
}
