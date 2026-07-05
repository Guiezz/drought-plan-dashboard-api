package scheduler

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestScheduler_ExecutaListerEUpdater(t *testing.T) {
	lister := func() ([]uint, error) {
		return []uint{1, 2}, nil
	}

	var calls atomic.Int32
	updater := func(id uint) (int, error) {
		calls.Add(1)
		return 1, nil
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	Start(ctx, 100*time.Millisecond, 10*time.Millisecond, 10*time.Millisecond, lister, updater)

	time.Sleep(80 * time.Millisecond)
	cancel()
	time.Sleep(30 * time.Millisecond)

	assert.GreaterOrEqual(t, calls.Load(), int32(2), "deveria ter chamado updater para cada ID")
}

func TestScheduler_ListerComErroNaoPanica(t *testing.T) {
	lister := func() ([]uint, error) {
		return nil, errors.New("erro no banco")
	}

	updater := func(id uint) (int, error) {
		return 0, nil
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	Start(ctx, 100*time.Millisecond, 10*time.Millisecond, 10*time.Millisecond, lister, updater)
	time.Sleep(80 * time.Millisecond)
	cancel()
}

func TestScheduler_UpdaterComErroContinuaProximo(t *testing.T) {
	var calls atomic.Int32

	lister := func() ([]uint, error) {
		return []uint{1, 2, 3}, nil
	}

	updater := func(id uint) (int, error) {
		calls.Add(1)
		if id == 2 {
			return 0, errors.New("erro na atualização")
		}
		return 1, nil
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	Start(ctx, 200*time.Millisecond, 10*time.Millisecond, 10*time.Millisecond, lister, updater)
	time.Sleep(100 * time.Millisecond)
	cancel()
	time.Sleep(30 * time.Millisecond)

	assert.Equal(t, int32(3), calls.Load(), "deveria ter chamado updater para todos os IDs mesmo com erro no meio")
}

func TestScheduler_ContextCancelParaTicker(t *testing.T) {
	var calls atomic.Int32

	lister := func() ([]uint, error) {
		calls.Add(1)
		return []uint{1}, nil
	}

	updater := func(id uint) (int, error) {
		return 1, nil
	}

	ctx, cancel := context.WithCancel(context.Background())

	Start(ctx, 30*time.Millisecond, 10*time.Millisecond, 10*time.Millisecond, lister, updater)
	time.Sleep(50 * time.Millisecond)

	initialCalls := calls.Load()
	assert.Greater(t, initialCalls, int32(0), "deveria ter executado ao menos uma vez")

	cancel()
	time.Sleep(80 * time.Millisecond)

	afterCancel := calls.Load()
	assert.Equal(t, initialCalls, afterCancel, "após cancelar, não deve executar novamente")
}
