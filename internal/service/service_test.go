package service

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/lovelaze/nebula-sync/internal/config"
	syncmock "github.com/lovelaze/nebula-sync/internal/mocks/sync"
	"github.com/lovelaze/nebula-sync/internal/pihole/model"
)

func TestRun_full(t *testing.T) {
	conf := config.Config{
		Primary:  model.PiHole{},
		Replicas: []model.PiHole{},
		Sync: &config.Sync{
			FullSync: true,
			Cron:     nil,
		},
	}

	target := syncmock.NewTarget(t)
	callback := syncmock.NewCallback(t)
	target.On("FullSync", conf.Sync).Return(nil)
	callback.On("OnSuccess").Return(nil)

	service := NewService(target, conf, callback)

	err := service.Run()
	require.NoError(t, err)

	target.AssertCalled(t, "FullSync", conf.Sync)
}

func TestRun_selective(t *testing.T) {
	conf := config.Config{
		Primary:  model.PiHole{},
		Replicas: []model.PiHole{},
		Sync: &config.Sync{
			FullSync: false,
			Cron:     nil,
		},
	}

	target := syncmock.NewTarget(t)
	callback := syncmock.NewCallback(t)
	target.On("SelectiveSync", conf.Sync).Return(nil)
	callback.On("OnSuccess").Return(nil)

	service := NewService(target, conf, callback)

	err := service.Run()
	require.NoError(t, err)

	target.AssertCalled(t, "SelectiveSync", conf.Sync)
}

func TestRun_webhook_success(t *testing.T) {
	conf := config.Config{
		Primary:  model.PiHole{},
		Replicas: []model.PiHole{},
		Sync: &config.Sync{
			FullSync: false,
			Cron:     nil,
		},
	}

	target := syncmock.NewTarget(t)
	callback := syncmock.NewCallback(t)

	target.On("SelectiveSync", conf.Sync).Return(nil)
	callback.On("OnSuccess").Return(nil)

	service := NewService(target, conf, callback)

	err := service.Run()
	require.NoError(t, err)

	target.AssertCalled(t, "SelectiveSync", conf.Sync)
	callback.AssertCalled(t, "OnSuccess")
	callback.AssertNotCalled(t, "OnFailure")
}

func TestRun_webhook_failure(t *testing.T) {
	conf := config.Config{
		Primary:  model.PiHole{},
		Replicas: []model.PiHole{},
		Sync: &config.Sync{
			FullSync: false,
			Cron:     nil,
		},
	}

	syncErr := errors.New("sync failed")
	target := syncmock.NewTarget(t)
	callback := syncmock.NewCallback(t)

	target.On("SelectiveSync", conf.Sync).Return(syncErr)
	callback.On("OnFailure", syncErr).Return(nil)

	service := NewService(target, conf, callback)

	err := service.Run()
	require.ErrorIs(t, err, syncErr)

	target.AssertCalled(t, "SelectiveSync", conf.Sync)
	callback.AssertCalled(t, "OnFailure", syncErr)
	callback.AssertNotCalled(t, "OnSuccess")
}

func TestRun_webhook_error_does_not_affect_result(t *testing.T) {
	conf := config.Config{
		Primary:  model.PiHole{},
		Replicas: []model.PiHole{},
		Sync: &config.Sync{
			FullSync: true,
			Cron:     nil,
		},
	}

	target := syncmock.NewTarget(t)
	callback := syncmock.NewCallback(t)

	target.On("FullSync", conf.Sync).Return(nil)
	callback.On("OnSuccess").Return(nil)

	service := NewService(target, conf, callback)

	err := service.Run()
	require.NoError(t, err)

	target.AssertCalled(t, "FullSync", conf.Sync)
	callback.AssertCalled(t, "OnSuccess")
}
