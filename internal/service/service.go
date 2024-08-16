package service

import (
	"fmt"
	"github.com/lovelaze/nebula-sync/internal/config"
	"github.com/lovelaze/nebula-sync/internal/pihole"
	"github.com/lovelaze/nebula-sync/internal/sync"
	"github.com/lovelaze/nebula-sync/version"
	"github.com/robfig/cron/v3"
	"github.com/rs/zerolog/log"
)

type Service struct {
	target sync.Target
	conf   config.Config
}

func Init() (*Service, error) {
	conf := config.Config{}
	if err := conf.Load(); err != nil {
		return nil, err
	}

	primary := pihole.NewClient(conf.Primary)
	var rs []pihole.Client
	for _, replica := range conf.Replicas {
		rs = append(rs, pihole.NewClient(replica))
	}

	return &Service{
		target: sync.NewTarget(primary, rs),
		conf:   conf,
	}, nil
}

func (service *Service) Run() error {
	log.Info().Msgf("Starting nebula-sync v%s", version.Version)
	log.Debug().Msgf("Settings cron=%v, fullsync=%v, syncsettings=%v", service.conf.Cron, service.conf.FullSync, service.conf.SyncSettings)

	if service.conf.Cron == nil {
		return service.doSync(service.target)
	} else {
		return service.startCron(func() {
			if err := service.doSync(service.target); err != nil {
				log.Error().Err(err).Msg("sync failed")
			}
		})
	}
}

func (service *Service) doSync(t sync.Target) (err error) {
	if service.conf.FullSync {
		err = t.FullSync()
	} else {
		err = t.ManualSync(service.conf.SyncSettings)
	}

	if err != nil {
		return fmt.Errorf("sync failed: %w", err)
	}

	log.Info().Msg("Sync complete")
	return err
}

func (service *Service) startCron(cmd func()) error {
	cron := cron.New()

	if _, err := cron.AddFunc(*service.conf.Cron, cmd); err != nil {
		return fmt.Errorf("failed to start cron job: %w", err)
	}

	cron.Run()
	return nil
}

func (service *Service) Target() sync.Target {
	return service.target
}
