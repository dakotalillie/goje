package goje

type QueueItem struct {
	Cancelled  bool `json:"cancelled"`
	Executable struct {
		Number int `json:"number"`
	} `json:"executable"`
}
