package model1

import (
	"context"
	"log/slog"
	"sync"

	"github.com/derailed/k9s/internal/slogs"
)

type jobFn func(ctx context.Context) error

type WorkerPool struct {
	semC     chan struct{}
	errC     chan error
	ctx      context.Context
	cancelFn context.CancelFunc
	mx       sync.RWMutex
	wg       sync.WaitGroup
	wge      sync.WaitGroup
	errs     []error
}

func NewWorkerPool(ctx context.Context, size int) *WorkerPool {
	_, cancelFn := context.WithCancel(ctx)

	p := WorkerPool{
		semC:     make(chan struct{}, size),
		errC:     make(chan error, 1),
		cancelFn: cancelFn,
		ctx:      ctx,
	}

	p.wge.Add(1)
	go func(wg *sync.WaitGroup) {
		defer wg.Done()
		for err := range p.errC {
			if err != nil {
				p.mx.Lock()
				p.errs = append(p.errs, err)
				p.mx.Unlock()
			}
		}
	}(&p.wge)

	return &p
}

func (p *WorkerPool) Add(job jobFn) {
	p.semC <- struct{}{}
	p.wg.Add(1)
	go func(ctx context.Context, wg *sync.WaitGroup, semC <-chan struct{}, errC chan<- error) {
		defer func() {
			<-semC
			wg.Done()
		}()
		if err := job(ctx); err != nil {
			slog.Error("Worker error", slogs.Error, err)
			errC <- err
		}
	}(p.ctx, &p.wg, p.semC, p.errC)
}

func (p *WorkerPool) Drain() []error {
	if p.cancelFn != nil {
		p.cancelFn()
		p.cancelFn = nil
	}
	p.wg.Wait()
	close(p.semC)
	close(p.errC)
	p.wge.Wait()

	p.mx.RLock()
	defer p.mx.RUnlock()
	return p.errs
}
