package redis

import (
	"context"
	"time"
)

type Lock interface {
	Lock(ctx context.Context, key, user string, expireTime time.Duration) bool
	Unlock(ctx context.Context, key, user string) bool
	LockedByAnyone(ctx context.Context, key string) bool
}
