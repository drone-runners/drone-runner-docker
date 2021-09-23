package livelog

import "time"

// Line represents a line in the logs.
type Line struct {
	Level     string            `json:"level"`
	Number    int               `json:"pos"`
	Message   string            `json:"out"`
	Timestamp time.Time         `json:"time"`
	Args      map[string]string `json:"args"`
}