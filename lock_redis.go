package redis

import (
	"context"
	"time"
)

const (
	redisLockPrefix = "redisLockPrefix:"
	delLuaScript    = "if redis.call('get',KEYS[1]) == ARGV[1] then return redis.call('del',KEYS[1]) else return 0 end"
)

var script = NewScript(delLuaScript)

type RedisLock struct {
	client RedisClient
}

func NewRedisLock(cli RedisClient) *RedisLock {
	return &RedisLock{client: cli}
}

func (r *RedisLock) Lock(ctx context.Context, key, user string, expireTime time.Duration) bool {
	if expireTime <= 0 {
		return false
	}

	lock0, _ := r.lock0(ctx, key, user, expireTime)

	return lock0
}

func (r *RedisLock) lock0(ctx context.Context, key, user string, expireTime time.Duration) (bool, error) {
	result, err := r.client.SetNX(ctx, getRealKey(key), user, expireTime).Result()
	if err != nil {
		return false, err
	}
	return result, nil
}

func (r *RedisLock) Unlock(ctx context.Context, key, user string) bool {
	result, _ := r.unlock0(ctx, key, user)
	return result
}

func (r *RedisLock) unlock0(ctx context.Context, key, user string) (bool, error) {
	res, err := script.Run(ctx, r.client, []string{key}, user).Result()
	if err != nil {
		return false, err
	}
	i := res.(int64)
	if i == 0 {
		return false, nil
	}
	return true, nil
}

func (r *RedisLock) LockedByAnyone(ctx context.Context, key string) bool {
	res, _ := r.lockedByAnyone0(ctx, key)
	return res
}

func (r *RedisLock) lockedByAnyone0(ctx context.Context, key string) (bool, error) {
	result, err := r.client.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return result > 0, nil
}

func getRealKey(key string) string {
	return redisLockPrefix + key
}
