package client

import (
	"context"
	"sync"

	"github.com/hibiken/asynq"
)

type ctxKey int

const (
	_ ctxKey = iota
	asyncQCtxKey
)

var (
	globalClient *asynq.Client
	globalMu     sync.RWMutex
)

// GetClient returns the global Client, which can be reconfigured with SetClient.
// It's safe for concurrent use.
func GetClient(ctx context.Context) *asynq.Client {
	c := ctx.Value(asyncQCtxKey)
	if c != nil {
		client, ok := c.(*asynq.Client)
		if !ok {
			return nil
		}

		return client
	}

	globalMu.RLock()
	client := globalClient
	globalMu.RUnlock()

	return client
}

// SetClient replaces the global Client, and returns a
// function to restore the original value. It's safe for concurrent use.
func SetClient(client *asynq.Client) func() {
	globalMu.Lock()
	prev := globalClient
	globalClient = client
	globalMu.Unlock()
	return func() { SetClient(prev) }
}
