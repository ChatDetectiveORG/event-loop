package main

import (
	"context"
	"log"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/ChatDetectiveORG/event-loop/src/application"
	"github.com/ChatDetectiveORG/event-loop/src/infrastructure/config"
	"github.com/ChatDetectiveORG/event-loop/src/infrastructure/redis"
	"github.com/ChatDetectiveORG/shared/ratelimit"
)

func main() {
	cfg, err := config.FetchConfig()
	if !err.IsNil() {
		log.Fatal(err.JSON())
	}

	if initErr := redis.InitRedis(cfg); !initErr.IsNil() {
		log.Fatal(initErr.JSON())
	}
	pool, poolErr := redis.GetPool(cfg)
	if !poolErr.IsNil() {
		log.Fatal(poolErr.JSON())
	}
	ratelimit.SetPool(pool)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	wg := &sync.WaitGroup{}
	application.Run(ctx, cfg, wg)

	log.Println("event-loop started")
	<-ctx.Done()
	log.Println("event-loop shutdown signal received")

	waitCh := make(chan struct{})
	go func() {
		wg.Wait()
		close(waitCh)
	}()
	select {
	case <-waitCh:
	case <-time.After(30 * time.Second):
		log.Println("event-loop wait timeout, exiting")
	}
}
