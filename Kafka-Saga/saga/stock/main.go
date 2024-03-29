package main

import (
	"kafka-saga/saga/stock/dto"
	"kafka-saga/saga/stock/handlers"
	"log"
	"time"

	"github.com/Shopify/sarama"
	"golang.org/x/net/context"
	"kafka-saga/saga/consts"
)

type Store struct {
	data                   *dto.Map
	producer               sarama.SyncProducer
	incomeConsumer         *handlers.IncomeHandler
	resetConsumer          *handlers.ResetHandler
	resetBillStockConsumer *handlers.BillResetHandler
	confirmBillStConsumer  *handlers.BillConfirmStockHandler
}

func NewStock(ctx context.Context) (*Store, error) {
	Data := dto.NewMap()

	cfg := sarama.NewConfig()
	cfg.Producer.Return.Successes = true
	syncProducer, err := sarama.NewSyncProducer(consts.Brokers, cfg)
	if err != nil {
		return nil, err
	}

	reserve, err := sarama.NewConsumerGroup(consts.Brokers, "store", cfg)
	if err != nil {
		return nil, err
	}
	iHandler := &handlers.IncomeHandler{
		P:    syncProducer,
		Data: Data,
	}
	go func() {
		for {
			err := reserve.Consume(ctx, []string{"order_send"}, iHandler)
			if err != nil {
				log.Printf("reserve consumer error: %v", err)
				time.Sleep(time.Second * 5)
			}
			log.Printf("reserve consumer done")

		}
	}()

	// receiving resets from shop
	reset, err := sarama.NewConsumerGroup(consts.Brokers, "stockReset", cfg)
	if err != nil {
		return nil, err
	}
	rHandler := &handlers.ResetHandler{
		P:    syncProducer,
		Data: Data,
	}
	go func() {
		for {
			err := reset.Consume(ctx, []string{"shop_order_reset"}, rHandler)
			log.Printf("order reset")
			if err != nil {
				log.Printf("reset order error: %v", err)
				time.Sleep(time.Second * 5)
			}
		}
	}()

	// receiving resets from billing
	resetB, err := sarama.NewConsumerGroup(consts.Brokers, "billResetStock", cfg)
	if err != nil {
		return nil, err
	}
	rbStHandler := &handlers.BillResetHandler{
		P:    syncProducer,
		Data: Data,
	}
	go func() {
		for {
			err := resetB.Consume(ctx, []string{"bill_reset_stock"}, rbStHandler)
			log.Printf("billing order reset")
			if err != nil {
				log.Printf("reset consumer error: %v", err)
				time.Sleep(time.Second * 5)
			}
		}
	}()

	// receiving confirmation from billing
	confirm, err := sarama.NewConsumerGroup(consts.Brokers, "billConfirmedStock", cfg)
	if err != nil {
		return nil, err
	}
	bcStHandler := &handlers.BillConfirmStockHandler{
		P: syncProducer,
	}
	go func() {
		for {
			err := confirm.Consume(ctx, []string{"bill_confirmed_stock"}, bcStHandler)
			log.Printf("bill confirmed")
			if err != nil {
				log.Printf("bill confirm consume error: %v", err)
				time.Sleep(time.Second * 5)
			}
		}
	}()

	return &Store{
		data:                   dto.NewMap(),
		producer:               syncProducer,
		incomeConsumer:         iHandler,
		resetConsumer:          rHandler,
		resetBillStockConsumer: rbStHandler,
		confirmBillStConsumer:  bcStHandler,
	}, nil
}

func main() {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	_, err := NewStock(ctx)
	if err != nil {
		log.Fatalf("NewStock: %v", err)
	}
	<-ctx.Done()
}
