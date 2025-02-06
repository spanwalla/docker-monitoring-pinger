package entity

import "time"

type Report struct {
	ContainerId string    `json:"id"`
	Ip          string    `json:"ip"`
	Latency     int       `json:"latency_ms"`
	Status      string    `json:"status"`
	State       string    `json:"state"`
	Timestamp   time.Time `json:"timestamp"`
}
