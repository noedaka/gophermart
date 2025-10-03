package worker

import (
	"context"
	"gophermart/internal/client"
	"gophermart/internal/model"
	"gophermart/internal/repository"
	"log"
	"sync"
	"sync/atomic"
	"time"
)

type Worker struct {
	repo         *repository.Repository
	client       *client.Client
	workers      int
	batchSize    int
	pollInterval time.Duration
	
	// Механизм координированной паузы
	pauseUntil   atomic.Value
	jobs         chan model.Order
}

func NewWorker(repo *repository.Repository, client *client.Client, workers int) *Worker {
	w := &Worker{
		repo:         repo,
		client:       client,
		workers:      workers,
		batchSize:    10,
		pollInterval: 1 * time.Second,
		jobs:         make(chan model.Order, workers*2),
	}
	w.pauseUntil.Store(time.Time{})
	return w
}

func (w *Worker) Start(ctx context.Context) {
	var wg sync.WaitGroup

	for i := 0; i < w.workers; i++ {
		wg.Add(1)
		go w.worker(ctx, &wg, i)
	}

	wg.Add(1)
	go w.scheduler(ctx, &wg)

	log.Printf("Accrual worker pool started with %d workers", w.workers)

	wg.Wait()
	log.Println("Accrual worker pool stopped")
}

func (w *Worker) scheduler(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	ticker := time.NewTicker(w.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			w.fetchAndDispatchOrders(ctx)
		}
	}
}

func (w *Worker) fetchAndDispatchOrders(ctx context.Context) {
	if w.isPaused() {
		return
	}

	orders, err := w.repo.GetOrdersForProcessing(ctx, w.batchSize)
	if err != nil {
		log.Printf("Error getting orders for processing: %v", err)
		return
	}

	for _, order := range orders {
		if w.isPaused() {
			return
		}

		select {
		case w.jobs <- order:
		case <-ctx.Done():
			return
		default:
			// Если канал заполнен, пропускаем остальные заказы до следующей итерации
			return
		}
	}
}

func (w *Worker) worker(ctx context.Context, wg *sync.WaitGroup, workerID int) {
	defer wg.Done()

	for {
		// Проверяем паузу перед получением задачи
		if pauseUntil := w.getPauseUntil(); !pauseUntil.IsZero() {
			if !w.waitForPause(ctx, pauseUntil, workerID) {
				return 
			}
			continue 
		}

		select {
		case <-ctx.Done():
			return
		case order, ok := <-w.jobs:
			if !ok {
				return
			}
			w.processOrder(ctx, order, workerID)
		}
	}
}

func (w *Worker) processOrder(ctx context.Context, order model.Order, workerID int) {
	if w.isPaused() {
		return
	}

	log.Printf("Worker %d processing order %s", workerID, order.Number)

	orderInfo, err := w.client.GetOrderInfo(order.Number)
	if err != nil {
		if tooManyRequests, ok := err.(*client.TooManyRequestsError); ok {
			log.Printf("Worker %d got 429 error, pausing ALL workers for %v", workerID, tooManyRequests.RetryAfter)
			w.pauseAllWorkers(tooManyRequests.RetryAfter)
			return
		}
		log.Printf("Worker %d error getting order info for %s: %v", workerID, order.Number, err)
		return
	}

	// Обновляем заказ только если статус изменился или изменилось начисление
	if orderInfo.Status != order.Status || orderInfo.Accrual != order.Accrual {
		err = w.repo.UpdateOrderAccrual(ctx, *orderInfo)
		if err != nil {
			log.Printf("Worker %d error updating order %s: %v", workerID, order.Number, err)
		} else {
			log.Printf("Worker %d order %s updated to status %s", workerID, order.Number, orderInfo.Status)
		}
	}
}

func (w *Worker) pauseAllWorkers(retryAfter time.Duration) {
	pauseUntil := time.Now().Add(retryAfter)
	w.setPauseUntil(pauseUntil)
	log.Printf("ALL workers paused until %v", pauseUntil.Format("15:04:05.000"))
}

func (w *Worker) waitForPause(ctx context.Context, pauseUntil time.Time, workerID int) bool {
	now := time.Now()
	if now.After(pauseUntil) {
		w.setPauseUntil(time.Time{})
		return true
	}

	sleepDuration := pauseUntil.Sub(now)
	log.Printf("Worker %d waiting for pause to end: %v", workerID, sleepDuration)

	timer := time.NewTimer(sleepDuration)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		// Пауза завершена, сбрасываем состояние
		w.setPauseUntil(time.Time{})
		log.Printf("Worker %d pause completed, resuming work", workerID)
		return true
	}
}


func (w *Worker) setPauseUntil(pauseUntil time.Time) {
	w.pauseUntil.Store(pauseUntil)
}

func (w *Worker) getPauseUntil() time.Time {
	return w.pauseUntil.Load().(time.Time)
}

func (w *Worker) isPaused() bool {
	pauseUntil := w.getPauseUntil()
	return !pauseUntil.IsZero() && time.Now().Before(pauseUntil)
}