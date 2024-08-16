package service

import (
	"github.com/lovelaze/nebula-sync/internal/config"
	syncmock "github.com/lovelaze/nebula-sync/internal/mocks/sync"
	"github.com/lovelaze/nebula-sync/internal/pihole/model"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestRun_full(t *testing.T) {
	conf := config.Config{
		Primary:      model.PiHole{},
		Replicas:     []model.PiHole{},
		FullSync:     true,
		Cron:         nil,
		SyncSettings: nil,
	}

	target := syncmock.NewTarget(t)
	target.On("FullSync").Return(nil)

	service := Service{
		target: target,
		conf:   conf,
	}

	err := service.Run()
	require.NoError(t, err)

	target.AssertCalled(t, "FullSync")
}

func TestRun_manual(t *testing.T) {
	conf := config.Config{
		Primary:      model.PiHole{},
		Replicas:     []model.PiHole{},
		FullSync:     false,
		Cron:         nil,
		SyncSettings: nil,
	}

	target := syncmock.NewTarget(t)
	target.On("ManualSync", (*config.SyncSettings)(nil)).Return(nil)

	service := Service{
		target: target,
		conf:   conf,
	}

	err := service.Run()
	require.NoError(t, err)

	target.AssertCalled(t, "ManualSync", (*config.SyncSettings)(nil))
}
