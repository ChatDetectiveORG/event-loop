// Package application wires up event-loop background workers. Today this is just the rate-limit
// dispatcher; future cleanup / scheduled-mailing workers live alongside it without changing main.
package application

import (
	"context"
	"log"
	"sync"

	"github.com/ChatDetectiveORG/event-loop/src/application/leveltermination"
	"github.com/ChatDetectiveORG/event-loop/src/infrastructure/config"
	"github.com/ChatDetectiveORG/event-loop/src/infrastructure/referral"
	"github.com/ChatDetectiveORG/shared/ratelimit"
)

// Run starts every background worker owned by event-loop and blocks until ctx is done.
func Run(ctx context.Context, cfg *config.Config, wg *sync.WaitGroup) {
	startRateLimitDispatcher(ctx, cfg, wg)
	startReferralRewardManager(ctx, cfg, wg)
	startLevelTermination(ctx, cfg, wg)

	// Hook for future workers: cleanup, scheduled mailings, etc.
	// startCleanupWorker(ctx, cfg, wg)
	// startScheduledMailingsWorker(ctx, cfg, wg)
}

func startRateLimitDispatcher(ctx context.Context, cfg *config.Config, wg *sync.WaitGroup) {
	rate := cfg.RuntimeConfig.TokenBucketRate
	log.Printf("event-loop: starting rate limit dispatcher rate=%d/s", rate)

	wg.Add(1)
	go func() {
		defer wg.Done()
		ratelimit.StartDispatcher(ctx, rate)
		<-ctx.Done()
	}()
}

func startReferralRewardManager(ctx context.Context, cfg *config.Config, wg *sync.WaitGroup) {
	rate := cfg.RuntimeConfig.ReferralRewardManagerInterval
	log.Printf("event-loop: starting referral reward manager interval=%s", rate)

	wg.Add(1)
	go func() {
		defer wg.Done()
		referral.StartReferralRewardManagerLoop(ctx, rate)
		<-ctx.Done()
	}()

}

func startLevelTermination(ctx context.Context, cfg *config.Config, wg *sync.WaitGroup) {
	interval := cfg.RuntimeConfig.LevelTerminationInterval
	log.Printf("event-loop: starting level termination loop interval=%s", interval)

	wg.Add(1)
	go func() {
		defer wg.Done()
		leveltermination.StartLevelTerminationLoop(ctx, interval, cfg)
		<-ctx.Done()
	}()
}
