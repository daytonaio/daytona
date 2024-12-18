package log

import (
    "encoding/json"
    "time"
)

// LogPosition tracks the reading position in logs
type LogPosition struct {
    Offset    int64     `json:"offset"`
    Timestamp time.Time `json:"timestamp"`
    LastLine  string    `json:"last_line,omitempty"`
}

func (p *LogPosition) Marshal() string {
    data, _ := json.Marshal(p)
    return string(data)
}

func UnmarshalPosition(data string) (*LogPosition, error) {
    var pos LogPosition
    if err := json.Unmarshal([]byte(data), &pos); err != nil {
        return nil, err
    }
    return &pos, nil
}