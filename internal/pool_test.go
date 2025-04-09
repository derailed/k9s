package internal_test

import (
	"context"
	"fmt"
	"sync/atomic"
	"testing"

	"github.com/derailed/k9s/internal"
	"github.com/stretchr/testify/assert"
)

func TestWorkerPoolPlain(t *testing.T) {
	p := internal.NewWorkerPool(context.Background(), 2)

	var c atomic.Int32
	for range 10 {
		p.Add(func(ctx context.Context) error {
			select {
			case <-ctx.Done():
				fmt.Println("Worker canceled")
				return nil
			default:
				c.Add(1)
				return nil
			}
		})
	}
	errs := p.Drain()
	assert.Equal(t, 10, int(c.Load()))
	assert.Empty(t, errs)
}

func TestWorkerPoolWithError(t *testing.T) {
	ctx := context.Background()
	p := internal.NewWorkerPool(ctx, 2)

	var c atomic.Int32
	for i := range 10 {
		p.Add(func(ctx context.Context) error {
			select {
			case <-ctx.Done():
				fmt.Println("Worker canceled")
				return nil
			default:
				if i%2 == 0 {
					return fmt.Errorf("BOOM%d", i)
				}
				c.Add(1)
				return nil
			}
		})
	}
	errs := p.Drain()
	assert.Equal(t, 5, int(c.Load()))
	assert.Len(t, errs, 5)
}
