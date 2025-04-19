package service

import (
	"fmt"

	"github.com/robfig/cron/v3"
	"github.com/rs/zerolog/log"

	"github.com/lovelaze/nebula-sync/internal/api"
	"github.com/lovelaze/nebula-sync/internal/config"
	"github.com/lovelaze/nebula-sync/internal/pihole"
	"github.com/lovelaze/nebula-sync/internal/sync"
	"github.com/lovelaze/nebula-sync/internal/sync/retry"
	"github.com/lovelaze/nebula-sync/internal/webhook"
	"github.com/lovelaze/nebula-sync/version"
)

type Service struct {
	target    sync.Target
	conf      config.Config
	callbacks []sync.Callback
	State     *sync.State
}

func NewService(target sync.Target, conf config.Config, callbacks ...sync.Callback) *Service {
	state := sync.NewState()
	cbs := append([]sync.Callback{state}, callbacks...)

	return &Service{
		target:    target,
		conf:      conf,
		callbacks: cbs,
		State:     state,
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

	if conf.API.Enabled && conf.Sync.Cron != nil {
		server := api.NewServer(service.State)
		server.Start()
	}

	return service, nil
}

func (service *Service) Run() error {
	log.Info().Msgf("Starting nebula-sync %s", version.Version)
	log.Debug().Str("config", service.conf.String()).Msgf("Settings")

	if err := service.sync(service.target); err != nil {
		return err
	}

	if service.conf.Sync.Cron != nil {
		return service.startCron(func() {
			if err := service.sync(service.target); err != nil {
				log.Error().Err(err).Msg("Sync failed")
			}
		})
	}

	return nil
}

func (service *Service) sync(t sync.Target) error {
	var err error
	if service.conf.Sync.FullSync {
		err = t.FullSync(service.conf.Sync)
	} else {
		err = t.SelectiveSync(service.conf.Sync)
	}

	service.runCallbacks(err)

	if err == nil {
		log.Info().Msg("Sync completed")
	}

	return err
}

func (service *Service) runCallbacks(syncError error) {
	for _, callback := range service.callbacks {
		if syncError != nil {
			callback.OnFailure(syncError)
		} else {
			callback.OnSuccess()
		}
	}
}

func (service *Service) startCron(cmd func()) error {
	cron := cron.New()

	if _, err := cron.AddFunc(*service.conf.Sync.Cron, cmd); err != nil {
		return fmt.Errorf("cron job: %w", err)
	}

	cron.Run()
	return nil
}
