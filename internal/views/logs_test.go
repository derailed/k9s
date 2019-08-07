package views

import (
	"context"
	"fmt"
	"sync"
	"testing"

	"github.com/derailed/k9s/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestUpdateLogs(t *testing.T) {
	v := newLogView("test", NewApp(config.NewConfig(ks{})), nil)

	var wg sync.WaitGroup
	wg.Add(1)
	c := make(chan string, 10)
	go func() {
		defer wg.Done()
		updateLogs(context.Background(), c, v, 10)
	}()

	for i := 0; i < 500; i++ {
		c <- fmt.Sprintf("log %d", i)
	}
	close(c)
	wg.Wait()

	assert.Equal(t, 500, v.logs.GetLineCount())
}
