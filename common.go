package goje

type BuildResult string

const (
	BuildResultSuccess BuildResult = "SUCCESS"
	BuildResultFailure BuildResult = "FAILURE"
)

type Build struct {
	Building bool        `json:"building"`
	Result   BuildResult `json:"result"`
	URL      string      `json:"url"`
}

type Job struct {
	Name  string `json:"name"`
	URL   string `json:"url"`
	Color string `json:"color"`
}

type PendingInputAction struct {
	ID string `json:"id"`
}

type QueueItem struct {
	Cancelled  bool                `json:"cancelled"`
	Executable QueueItemExecutable `json:"executable"`
}

type QueueItemExecutable struct {
	Number int `json:"number"`
}
