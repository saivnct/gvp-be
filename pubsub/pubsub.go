package pubsub

import (
	"context"
	"fmt"
	"gbb.go/gvp/dao"
	"log"
)

const (
	EVENT_SV_HEALTHCHECK = "SV_HealthCheck"
)

var Events = []string{EVENT_SV_HEALTHCHECK}

func StartSubscribe() {
	cache := dao.GetCache()
	subscriber := cache.RedisClient.Subscribe(context.Background(), Events...)

	go func() {
		for {
			msg, err := subscriber.ReceiveMessage(context.Background())
			if err != nil {
				log.Println("redis Subscribe error", err)
			}

			switch msg.Channel {
			case EVENT_SV_HEALTHCHECK:
				fmt.Println("redis: received - EVENT_SV_HEALTHCHECK")

				//TODO
				break
			default:
				fmt.Println("redis: unknown message", msg)
				break
			}

		}
	}()
}
