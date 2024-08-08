package service

import (
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

func NewService(conf config.Config) *Service {
	primary := pihole.NewClient(conf.Primary)
	var rs []pihole.Client
	for _, replica := range conf.Replicas {
		rs = append(rs, pihole.NewClient(replica))
	}

	return &Service{
		target: sync.NewTarget(primary, rs),
		conf:   conf,
	}
}

func (service *Service) Run() {
	log.Info().Msgf("Starting nebula-sync v%s", version.Version)
	log.Debug().Msgf("Settings cron=%v, fullsync=%v, syncsettings=%v", service.conf.Cron, service.conf.FullSync, service.conf.SyncSettings)

	if service.conf.Cron == nil {
		service.doSync(service.target)
	} else {
		service.startCron(func() {
			service.doSync(service.target)
		})
	}
}

func (service *Service) doSync(t sync.Target) {
	var err error
	if service.conf.FullSync {
		err = t.FullSync()
	} else {
		err = t.ManualSync(service.conf.SyncSettings)
	}

	if err != nil {
		log.Error().Err(err).Msgf("Sync failed")
		return
	}

	log.Info().Msg("Sync complete")
}

func (service *Service) startCron(cmd func()) {
	cron := cron.New()

	if _, err := cron.AddFunc(*service.conf.Cron, cmd); err != nil {
		log.Fatal().Err(err).Msgf("Failed to start cron: %s", *service.conf.Cron)
	}

	cron.Run()
}
