package worker

import (
	"context"
	"gophermart/internal/client"
	"gophermart/internal/model"
	"gophermart/internal/repository"
	"log"
	"sync"
	"time"
)

type Worker struct {
	repo         *repository.Repository
	client       *client.Client
	workers      int
	batchSize    int
	pollInterval time.Duration
}

func NewWorker(repo *repository.Repository, client *client.Client, workers int) *Worker {
	return &Worker{
		repo:         repo,
		client:       client,
		workers:      workers,
		batchSize:    10,
		pollInterval: 1 * time.Second,
	}
}

func (w *Worker) Start(ctx context.Context) {
	jobs := make(chan model.Order, w.workers*2)
	var wg sync.WaitGroup

	for i := 0; i < w.workers; i++ {
		wg.Add(1)
		go w.worker(ctx, &wg, jobs, i)
	}

	// Запускаем планировщик заданий
	wg.Add(1)
	go w.scheduler(ctx, &wg, jobs)

	log.Printf("Accrual worker pool started with %d workers", w.workers)

	// Ждем завершения всех горутин
	wg.Wait()
	log.Println("Accrual worker pool stopped")
}

func (w *Worker) scheduler(ctx context.Context, wg *sync.WaitGroup, jobs chan<- model.Order) {
	defer wg.Done()
	defer close(jobs)

	ticker := time.NewTicker(w.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			w.fetchAndDispatchOrders(ctx, jobs)
		}
	}
}

func (w *Worker) fetchAndDispatchOrders(ctx context.Context, jobs chan<- model.Order) {
	orders, err := w.repo.GetOrdersForProcessing(ctx, w.batchSize)
	if err != nil {
		log.Printf("Error getting orders for processing: %v", err)
		return
	}

	for _, order := range orders {
		select {
		case jobs <- order:
		case <-ctx.Done():
			return
		default:
			// Если канал заполнен, пропускаем остальные заказы до следующей итерации
			return
		}
	}
}

func (w *Worker) worker(ctx context.Context, wg *sync.WaitGroup, jobs <-chan model.Order, workerID int) {
	defer wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		case order, ok := <-jobs:
			if !ok {
				return
			}
			w.processOrder(ctx, order, workerID)
		}
	}
}

func (w *Worker) processOrder(ctx context.Context, order model.Order, workerID int) {
	log.Printf("Worker %d processing order %s", workerID, order.Number)

	orderInfo, err := w.client.GetOrderInfo(order.Number)
	if err != nil {
		if tooManyRequests, ok := err.(*client.TooManyRequestsError); ok {
			log.Printf("Worker %d got 429 error, sleeping for %v", workerID, tooManyRequests.RetryAfter)
			w.handleTooManyRequests(ctx, tooManyRequests.RetryAfter)
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

func (w *Worker) handleTooManyRequests(ctx context.Context, retryAfter time.Duration) {
	timer := time.NewTimer(retryAfter)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return
	case <-timer.C:
		// Время ожидания истекло, продолжаем работу
	}
}
