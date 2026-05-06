package crawler

import (
	"context"
	"time"
)

type ProgressTicker struct {
	ctx       context.Context
	done      chan struct{}
	ticker    *time.Ticker
	displayFn func()
}

func (t *ProgressTicker) startTicking() {
	go func() {
		for {
			select {
			case <-t.ctx.Done():
				t.ticker.Stop()
				return
			case <-t.done:
				t.ticker.Stop()
				return
			case <-t.ticker.C:
				t.displayFn()
			}
		}
	}()
}

func (t *ProgressTicker) Display() {
	t.displayFn()
}

func (t *ProgressTicker) Stop() {
	t.done <- struct{}{}
}

func NewTickerProgress(ctx context.Context, freq time.Duration, tickerDisplayFn func()) ProgressTicker {
	ticker := time.NewTicker(freq)

	doneCh := make(chan struct{}, 1)
	return ProgressTicker{
		ctx:       ctx,
		done:      doneCh,
		ticker:    ticker,
		displayFn: tickerDisplayFn,
	}
}
