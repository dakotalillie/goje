package goje

import "encoding/json"

type BuildResult string

const (
	BuildResultSuccess BuildResult = "SUCCESS"
)

type Build struct {
	Building bool              `json:"building"`
	Result   BuildResult       `json:"result"`
	Actions  []json.RawMessage `json:"actions"`
}
