package sender

import "github.com/spanwalla/docker-monitoring-pinger/internal/entity"

type Sender interface {
	Send(data []entity.Report) error
}
