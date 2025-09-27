package worker

import (
	"context"
	"gophermart/internal/client"
	"gophermart/internal/model"
	"gophermart/internal/repository"
	"log"
	"time"
)

type Worker struct {
	repo      *repository.Repository
	client    *client.Client
	interval  time.Duration
	batchSize int
}

func NewWorker(repo *repository.Repository, client *client.Client) *Worker {
	return &Worker{
		repo:      repo,
		client:    client,
		interval:  5 * time.Second,
		batchSize: 10,
	}
}

func (w *Worker) Start(ctx context.Context) {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	log.Println("Accrual worker started")

	for {
		select {
		case <-ctx.Done():
			log.Println("Accrual worker stopped")
			return
		case <-ticker.C:
			w.processOrders(ctx)
		}
	}
}

func (w *Worker) processOrders(ctx context.Context) {
	orders, err := w.repo.GetOrdersForProcessing(ctx, w.batchSize)
	if err != nil {
		log.Printf("Error getting orders for processing: %v", err)
		return
	}

	for _, order := range orders {
		select {
		case <-ctx.Done():
			return
		default:
			w.processOrder(ctx, order)
		}
	}
}

func (w *Worker) processOrder(ctx context.Context, order model.Order) {
	log.Printf("Processing order %s", order.Number)

	orderInfo, err := w.client.GetOrderInfo(order.Number)
	if err != nil {
		log.Printf("Error getting order info for %s: %v", order.Number, err)
		return
	}

	// Обновляем заказ только если статус изменился
	if orderInfo.Status != order.Status {
		err = w.repo.UpdateOrderAccrual(ctx, *orderInfo)
		if err != nil {
			log.Printf("Error updating order %s: %v", order.Number, err)
		} else {
			log.Printf("Order %s updated to status %s", order.Number, orderInfo.Status)
		}
	}
}
