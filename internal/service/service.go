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

	httpClient := conf.Client.NewHttpClient()

	primary := pihole.NewClient(conf.Primary, httpClient)
	var replicas []pihole.Client
	for _, replica := range conf.Replicas {
		replicas = append(replicas, pihole.NewClient(replica, httpClient))
	}

	return &Service{
		target: sync.NewTarget(primary, replicas, conf.Client),
		conf:   conf,
	}, nil
}

func (service *Service) Run() error {
	log.Info().Msgf("Starting nebula-sync %s", version.Version)
	log.Debug().Str("config", service.conf.String()).Msgf("Settings")

	if err := service.doSync(service.target); err != nil {
		return err
	}

	if service.conf.Sync.Cron != nil {
		return service.startCron(func() {
			if err := service.doSync(service.target); err != nil {
				log.Error().Err(err).Msg("Sync failed")
			}
		})
	}

	return nil
}

func (service *Service) doSync(t sync.Target) (err error) {
	if service.conf.Sync.FullSync {
		err = t.FullSync(service.conf.Sync)
	} else {
		err = t.SelectiveSync(service.conf.Sync)
	}

	if err != nil {
		return err
	}

	log.Info().Msg("Sync complete")
	return err
}

func (service *Service) startCron(cmd func()) error {
	cron := cron.New()

	if _, err := cron.AddFunc(*service.conf.Sync.Cron, cmd); err != nil {
		return fmt.Errorf("cron job: %w", err)
	}

	cron.Run()
	return nil
}
