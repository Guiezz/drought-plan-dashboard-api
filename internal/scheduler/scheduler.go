package scheduler

import (
	"context"
	"log"
	"time"
)

type ReservoirLister func() ([]uint, error)
type ReservoirUpdater func(id uint) (int, error)

func Start(ctx context.Context, interval time.Duration, lister ReservoirLister, updater ReservoirUpdater) {
	go func() {
		time.Sleep(10 * time.Second)

		runUpdate(lister, updater)

		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				runUpdate(lister, updater)
			case <-ctx.Done():
				log.Println("[Scheduler] Encerrando atualização automática da Funceme")
				return
			}
		}
	}()
}

func runUpdate(lister ReservoirLister, updater ReservoirUpdater) {
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
		time.Sleep(2 * time.Second)
	}

	log.Println("[Scheduler] Atualização concluída")
}
