package crawler

import (
	"context"
	"time"
)

type ProgressTicker struct {
	ctx       context.Context
	freqMs    int
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

func NewTickerProgress(ctx context.Context, freqMs int, tickerDisplayFn func()) ProgressTicker {
	ticker := time.NewTicker(time.Duration(freqMs) * time.Millisecond)

	doneCh := make(chan struct{})
	return ProgressTicker{
		ctx:       ctx,
		freqMs:    freqMs,
		done:      doneCh,
		ticker:    ticker,
		displayFn: tickerDisplayFn,
	}
}
