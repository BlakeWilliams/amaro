package job

import (
	"container/list"
	"context"
)

type MemoryStorage struct {
	queues map[string]*list.List
}

func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		queues: make(map[string]*list.List),
	}
}

func (ms *MemoryStorage) PushJob(ctx context.Context, queueName string, payload string) error {
	if _, ok := ms.queues[queueName]; !ok {
		ms.queues[queueName] = list.New()
	}

	ms.queues[queueName].PushBack(payload)

	return nil
}

func (ms *MemoryStorage) PopJob(ctx context.Context, queueName string) (string, error) {
	if _, ok := ms.queues[queueName]; !ok {
		return "", NothingToPopErr
	}

	value := ms.queues[queueName].Front()

	if value != nil {
		ms.queues[queueName].Remove(value)
		return value.Value.(string), nil
	}

	return "", NothingToPopErr
}
