package main

import (
	"encoding/json"
	"fmt"
	"golang.org/x/net/context"
	"kafka-saga/saga/shop/handlers"
	"log"
	"math/rand"
	"time"

	"github.com/Shopify/sarama"
	"kafka-saga/saga/consts"
)

type Shop struct {
	producer            sarama.SyncProducer
	resetConsumer       *handlers.ResetHandler
	reserveConsumer     *handlers.ReserveHandler
	resetBillConsumer   *handlers.ResetBillHandler
	confirmBillConsumer *handlers.BillConfirmHandler
}

func NewShop(ctx context.Context) (*Shop, error) {
	cfg := sarama.NewConfig()
	cfg.Producer.Return.Successes = true

	// sending order
	syncProducer, err := sarama.NewSyncProducer(consts.Brokers, cfg)
	if err != nil {
		log.Fatalf("sync kafka: %v", err)
		return nil, err
	}
	go func() {
		for {
			d := consts.Order{
				Id:   int(time.Now().UnixNano()),
				Data: time.Now().Format("order 150405.000"),
			}
			b, err := json.Marshal(d)
			if err != nil {
				log.Printf("marshal unsuccessful %v", err)
				continue
			}
			//par, off, err := syncProducer.SendMessage(&sarama.ProducerMessage{
			_, _, err = syncProducer.SendMessage(&sarama.ProducerMessage{
				Topic: "order_send",
				Key:   sarama.StringEncoder(fmt.Sprintf("%v", d.Id)),
				Value: sarama.ByteEncoder(b),
			})
			//log.Printf("order %v -> %v; %v", par, off, err)
			log.Printf("order %v sent; error: %v", d.Id, err)

			time.Sleep(time.Millisecond * 500)

			if rand.Intn(10) == 9 {
				_, _, err = syncProducer.SendMessage(&sarama.ProducerMessage{
					Topic: "shop_order_reset",
					Key:   sarama.StringEncoder(fmt.Sprintf("%v", d.Id)),
					Value: sarama.ByteEncoder(b),
				})
				//log.Printf("reset %v -> %v; %v", par, off, err)
				log.Printf("shop resets order %v; error: %v", d.Id, err)
			}
		}
	}()

	// receiving resets from stock
	reset, err := sarama.NewConsumerGroup(consts.Brokers, "shopReset", cfg)
	if err != nil {
		return nil, err
	}
	rHandler := &handlers.ResetHandler{
		P: syncProducer,
	}
	go func() {
		for {
			err := reset.Consume(ctx, []string{"stock_order_reset"}, rHandler)
			log.Printf("order reset")
			if err != nil {
				log.Printf("reset consumer error: %v", err)
				time.Sleep(time.Second * 5)
			}
		}
	}()

	// receiving resets from billing
	resetB, err := sarama.NewConsumerGroup(consts.Brokers, "billResetShop", cfg)
	if err != nil {
		return nil, err
	}
	rbHandler := &handlers.ResetBillHandler{
		P: syncProducer,
	}
	go func() {
		for {
			err := resetB.Consume(ctx, []string{"bill_reset_shop"}, rbHandler)
			log.Printf("billing order reset")
			if err != nil {
				log.Printf("reset consumer error: %v", err)
				time.Sleep(time.Second * 5)
			}
		}
	}()

	// receiving reserves from stock
	reserve, err := sarama.NewConsumerGroup(consts.Brokers, "shopReserve", cfg)
	if err != nil {
		return nil, err
	}
	rsHandler := &handlers.ReserveHandler{
		P: syncProducer,
	}
	go func() {
		for {
			err := reserve.Consume(ctx, []string{"order_reserved"}, rsHandler)
			log.Printf("order reserved")
			if err != nil {
				log.Printf("reserve order error: %v", err)
				time.Sleep(time.Second * 5)
			}
		}
	}()

	// receiving confirmation from billing
	confirm, err := sarama.NewConsumerGroup(consts.Brokers, "billConfirmed", cfg)
	if err != nil {
		return nil, err
	}
	bcHandler := &handlers.BillConfirmHandler{
		P: syncProducer,
	}
	go func() {
		for {
			err := confirm.Consume(ctx, []string{"bill_confirmed_shop"}, bcHandler)
			log.Printf("bill confirmed")
			if err != nil {
				log.Printf("bill confirm consume error: %v", err)
				time.Sleep(time.Second * 5)
			}
		}
	}()

	return &Shop{
		producer:            syncProducer,
		resetConsumer:       rHandler,
		reserveConsumer:     rsHandler,
		resetBillConsumer:   rbHandler,
		confirmBillConsumer: bcHandler,
	}, nil
}

func main() {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	_, err := NewShop(ctx)
	if err != nil {
		log.Fatalf("NewStock: %v", err)
	}
	<-ctx.Done()
}
