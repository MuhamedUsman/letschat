package common

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"
)

var (
	bt   *BackgroundTask
	once sync.Once
)

type BackgroundTask struct {
	wg     sync.WaitGroup
	ctx    context.Context
	cancel context.CancelFunc
}

func NewBackgroundTask() *BackgroundTask {
	ctx, cancel := context.WithCancel(context.Background())
	once.Do(func() {
		bt = &BackgroundTask{
			ctx:    ctx,
			cancel: cancel,
		}
	})
	return bt
}

func (bt *BackgroundTask) Run(fn func(shtdwnCtx context.Context)) {
	bt.wg.Add(1)
	go func() {
		defer bt.wg.Done()
		defer func() {
			if r := recover(); r != nil {
				slog.Error(fmt.Errorf("%v", r).Error())
			}
		}()
		fn(bt.ctx)
	}()
}

func (bt *BackgroundTask) Shutdown(timeout time.Duration) {
	bt.cancel()
	wait := make(chan struct{})
	go func() {
		bt.wg.Wait()
		close(wait)
	}()
	select {
	case <-wait:
		return
	case <-time.After(timeout):
		slog.Warn("Shutdown timeout, some tasks may not have finished")
	}
}
