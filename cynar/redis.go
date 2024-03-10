package cynar

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisClient struct {
	client  *redis.Client
	Timeout time.Duration
}

func NewRedisStorage(client *redis.Client) *RedisClient {
	return &RedisClient{
		client:  client,
		Timeout: 10 * time.Second,
	}
}

func (rc *RedisClient) PushJob(ctx context.Context, queueName string, payload string) error {
	_, err := rc.client.RPush(ctx, queueName, payload).Result()

	if err != nil {
		return fmt.Errorf("failed to push job to redis: %s", err)
	}

	return nil
}

func (rc *RedisClient) PopJob(ctx context.Context, queueName string) (string, error) {
	result, err := rc.client.BLPop(ctx, rc.Timeout, queueName).Result()

	if err != nil && errors.Is(err, redis.Nil) {
		return "", fmt.Errorf("redis queue is empty: %w", NothingToPopErr)
	} else if err != nil {
		return "", fmt.Errorf("failed to pop job from redis: %s", err)
	}

	payload := result[1]

	return payload, nil
}

var _ Storage = (*RedisClient)(nil)
