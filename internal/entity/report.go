package entity

import "time"

type Report struct {
	ContainerId string    `json:"id"`
	Ip          string    `json:"ip"`
	Latency     int       `json:"latency_ms"`
	Timestamp   time.Time `json:"timestamp"`
}
