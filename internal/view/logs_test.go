package view

import (
	"context"
	"fmt"
	"sync"
	"testing"

	"github.com/derailed/k9s/internal/config"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
)

func TestUpdateLogs(t *testing.T) {
	v := NewLog("test", NewApp(config.NewConfig(ks{})), nil)

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

// Helpers...

type ks struct{}

func (k ks) CurrentContextName() (string, error) {
	return "test", nil
}

func (k ks) CurrentClusterName() (string, error) {
	return "test", nil
}

func (k ks) CurrentNamespaceName() (string, error) {
	return "test", nil
}

func (k ks) ClusterNames() ([]string, error) {
	return []string{"test"}, nil
}

func (k ks) NamespaceNames(nn []v1.Namespace) []string {
	return []string{"test"}
}
