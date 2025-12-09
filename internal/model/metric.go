package model

import "time"

type Metric struct {
    Timestamp time.Time `json:"timestamp"`
    CPU       float64   `json:"cpu"`
    RPS       float64   `json:"rps"`
}
