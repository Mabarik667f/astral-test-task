package infrastructure

import (
	"context"
	"time"
)

type Cache interface {
	Set(k string, v any, ttl time.Duration)
	Get(k string) (any, bool)
	Delete(k string)
	StartCleanup(ctx context.Context, interval time.Duration)
}
