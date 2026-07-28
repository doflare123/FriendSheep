package events

import "github.com/robfig/cron/v3"

type cronPopularEventsScheduler struct {
	cron *cron.Cron
}

func NewCronPopularEventsScheduler() PopularEventsScheduler {
	return &cronPopularEventsScheduler{
		cron: cron.New(),
	}
}

func (s *cronPopularEventsScheduler) Schedule(spec string, job func()) error {
	_, err := s.cron.AddFunc(spec, job)
	return err
}

func (s *cronPopularEventsScheduler) Start() {
	s.cron.Start()
}

func (s *cronPopularEventsScheduler) Stop() {
	s.cron.Stop()
}
