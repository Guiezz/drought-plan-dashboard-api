package scheduler

import (
	"context"
	"log"
	"time"
)

type ReservoirLister func() ([]uint, error)
type ReservoirUpdater func(id uint) (int, error)

func Start(ctx context.Context, interval, initialDelay, perItemDelay time.Duration, lister ReservoirLister, updater ReservoirUpdater) {
	go func() {
		time.Sleep(initialDelay)

		runUpdate(lister, updater, perItemDelay)

		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				runUpdate(lister, updater, perItemDelay)
			case <-ctx.Done():
				log.Println("[Scheduler] Encerrando atualização automática da Funceme")
				return
			}
		}
	}()
}

func runUpdate(lister ReservoirLister, updater ReservoirUpdater, perItemDelay time.Duration) {
	log.Println("[Scheduler] Iniciando atualização dos dados da Funceme...")

	ids, err := lister()
	if err != nil {
		log.Printf("[Scheduler] Erro ao listar reservatórios: %v", err)
		return
	}

	for _, id := range ids {
		novos, err := updater(id)
		if err != nil {
			log.Printf("[Scheduler] Erro ao atualizar reservatório %d: %v", id, err)
		} else if novos > 0 {
			log.Printf("[Scheduler] Reservatório %d: %d novos registros", id, novos)
		}
		time.Sleep(perItemDelay)
	}

	log.Println("[Scheduler] Atualização concluída")
}
