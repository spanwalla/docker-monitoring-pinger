package collector

import "github.com/spanwalla/docker-monitoring-pinger/internal/entity"

type Collector interface {
	Collect() ([]entity.Report, error)
}
