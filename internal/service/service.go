package service

import (
	"fmt"
	"github.com/lovelaze/nebula-sync/internal/sync/retry"
	"github.com/lovelaze/nebula-sync/internal/webhook"

	"github.com/lovelaze/nebula-sync/internal/config"
	"github.com/lovelaze/nebula-sync/internal/pihole"
	"github.com/lovelaze/nebula-sync/internal/sync"
	"github.com/lovelaze/nebula-sync/version"
	"github.com/robfig/cron/v3"
	"github.com/rs/zerolog/log"
)

type Service struct {
	target    sync.Target
	conf      config.Config
	callbacks []sync.Callback
}

func NewService(target sync.Target, conf config.Config, callbacks ...sync.Callback) *Service {
	return &Service{
		target:    target,
		conf:      conf,
		callbacks: callbacks,
	}
}

func Init() (*Service, error) {
	conf := config.Config{}
	if err := conf.Load(); err != nil {
		return nil, err
	}

	httpClient := conf.Client.NewHTTPClient()
	retry.Init(conf.Client)

	primary := pihole.NewClient(conf.Primary, httpClient)
	var replicas []pihole.Client
	for _, replica := range conf.Replicas {
		replicas = append(replicas, pihole.NewClient(replica, httpClient))
	}

	webhookClient := webhook.NewClient(conf.Sync.WebhookSettings)

	target := sync.NewTarget(primary, replicas)
	service := NewService(target, conf, webhookClient)
	return service, nil
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

func (service *Service) doSync(t sync.Target) error {
	err := service.selectSyncMethod(t)

	if err != nil {
		for _, callback := range service.callbacks {
			callback.OnFailure(err)
		}
	} else {
		for _, callback := range service.callbacks {
			callback.OnSuccess()
		}
		log.Info().Msg("Sync completed")
	}

	return err
}

func (service *Service) selectSyncMethod(t sync.Target) error {
	if service.conf.Sync.FullSync {
		return t.FullSync(service.conf.Sync)
	}

	return t.SelectiveSync(service.conf.Sync)
}

func (service *Service) startCron(cmd func()) error {
	cron := cron.New()

	if _, err := cron.AddFunc(*service.conf.Sync.Cron, cmd); err != nil {
		return fmt.Errorf("cron job: %w", err)
	}

	cron.Run()
	return nil
}
