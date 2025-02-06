package scheduler

import (
	"context"
	"fmt"
	"github.com/robfig/cron/v3"
	log "github.com/sirupsen/logrus"
	"github.com/spanwalla/docker-monitoring-pinger/internal/collector"
	"github.com/spanwalla/docker-monitoring-pinger/internal/sender"
	"os"
	"os/signal"
	"syscall"
)

type TaskScheduler struct {
	cron *cron.Cron
}

func NewTaskScheduler() *TaskScheduler {
	return &TaskScheduler{
		cron: cron.New(),
	}
}

func (s *TaskScheduler) Start(collector collector.Collector, sender sender.Sender, cronSpec string) error {
	_, err := s.cron.AddFunc(cronSpec, func() {
		containers, err := collector.Collect()
		if err != nil {
			log.Errorf("TaskScheduler.Start - collector.Collect: %v", err)
			return
		}

		if err = sender.Send(containers); err != nil {
			log.Errorf("TaskScheduler.Start - sender.Send: %v", err)
			return
		}
		log.Debugf("TaskScheduler.Start - Sent report about %d containers", len(containers))
	})

	if err != nil {
		return fmt.Errorf("TaskScheduler.Start - cron.AddFunc: %v", err)
	}

	s.cron.Start()
	log.Infof("TaskScheduler.Start - scheduler started with cron spec: %s", cronSpec)
	return nil
}

func (s *TaskScheduler) Stop() {
	s.cron.Stop()
	log.Infof("TaskScheduler.Stop - scheduler stopped")
}

func (s *TaskScheduler) RunWithGracefulShutdown(collector collector.Collector, sender sender.Sender, cronSpec string) {
	if err := s.Start(collector, sender, cronSpec); err != nil {
		log.Fatalf("TaskScheduler.RunWithGracefulShutdown: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	select {
	case ch := <-interrupt:
		log.Infof("Task signal received: %s, stopping gracefully...", ch.String())
	case <-ctx.Done():
		log.Info("Context canceled, stopping gracefully...")
	}

	s.Stop()
	log.Info("Graceful shutdown complete")
}
