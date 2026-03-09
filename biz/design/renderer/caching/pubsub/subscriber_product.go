package pubsub

import (
	"context"
	"encoding/json"

	gcpPubSub "cloud.google.com/go/pubsub"
	productsvc "github.com/leeseika/cv-demo/biz/design/renderer/caching/service/product"
	"github.com/leeseika/cv-demo/pkg/constants"
	"github.com/leeseika/cv-demo/pkg/driver/pubsub"
	"github.com/leeseika/cv-demo/pkg/model/dto"
	"github.com/leeseika/cv-demo/pkg/threading"
	"github.com/rs/zerolog/log"
)

type productSubscriber struct {
	productSvc    productsvc.Product
	gcpSubscriber *pubsub.GooglePubsubSubscriber
}

func NewProductSubscriber(productSvc productsvc.Product, gcpSubscriber *pubsub.GooglePubsubSubscriber) Subscriber {
	return &productSubscriber{
		productSvc:    productSvc,
		gcpSubscriber: gcpSubscriber,
	}
}

func (ps *productSubscriber) Start(ctx context.Context) error {
	sub, err := ps.gcpSubscriber.CreateSubscription(ctx, constants.TopicProduct)
	if err != nil {
		return err
	}

	fn := func() {
		err = sub.Receive(ctx, func(ctx context.Context, msg *gcpPubSub.Message) {
			// 只在成功或不需要重试的情况下响应 ACK，需要重试的情况响应 Nack
			shouldAck := true
			defer func() {
				if shouldAck {
					msg.Ack()
				} else {
					msg.Nack()
				}
			}()

			var event dto.Event
			err := json.Unmarshal(msg.Data, &event)
			if err != nil {
				return
			}

			if event.Action != constants.ActionUpdated && event.Action != constants.ActionDeleted {
				return
			}

			p := event.Payload
			var payload *dto.EventPayloadProduct
			payload, ok := p.(*dto.EventPayloadProduct)
			if !ok {
				var payloadStruct dto.EventPayloadProduct
				payloadStruct, ok = p.(dto.EventPayloadProduct)
				if !ok {
					return
				}
				payload = &payloadStruct
			}

			id := payload.ProductID
			err = ps.productSvc.DeleteCacheByIDs(ctx, []string{id})
			if err != nil {
				// 响应 Nack，需要重试
				shouldAck = false
				return
			}
		})
		if err != nil {
			log.Err(err).Msg("failed to subscribe pub/sub message")
		}
	}

	threading.GoSafe(fn, "panic happened in product subscriber", nil)

	return nil
}

func (ps *productSubscriber) Stop(ctx context.Context) error {
	return ps.gcpSubscriber.Close()
}
