package job

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"reflect"
	"time"
)

var ErrNothingToPop = errors.New("nothing to pop")

type (
	// JobManager manages background publishing and running defined background
	// jobs.
	JobManager[T any] struct {
		storage    Storage
		jobContext T
		jobs       map[string]reflect.Type
		queueMap   map[reflect.Type]string
		Logger     *slog.Logger
	}

	// Storage is the interface that must be implemented by any storage
	// mechanism like redis, memory, etc. to push and pop jobs of of the given
	// queue.
	Storage interface {
		// Enqueue pushes a job onto the given queue.
		Enqueue(ctx context.Context, queueName string, payload string) error
		// Dequeue pops a job off of the given queue.
		Dequeue(ctx context.Context, queueName string) (string, error)
	}

	// A Job is any struct that can perform a background job after being
	// unmarshaled from a string.
	Job[T any] interface {
		PerformJob(T)
	}
)

// New creates a new background job manager using the given storage. The jobContext passed
// is passed to each background job's PerformJob method when run. This allows resources
// like database connections to be passed to background jobs.
func New[T any](storage Storage, t T) *JobManager[T] {
	return &JobManager[T]{
		storage:    storage,
		jobContext: t,
		jobs:       make(map[string]reflect.Type),
		queueMap:   make(map[reflect.Type]string),
		Logger:     slog.New(slog.NewJSONHandler(io.Discard, nil)),
	}
}

// Registers a new queue using the passed in name as the key, which is passed to
// the storge implementation.
func (m *JobManager[T]) RegisterQueue(name string, job Job[T]) {
	jobType := normalizeType(job)
	m.jobs[name] = jobType
	m.queueMap[jobType] = name
}

// Run creates a goroutine for each queue registered via `RegisterQueue` and
// uses the provided Storage to begin popping jobs off of the queue and
// calling PerformJob on each.
func (bm *JobManager[T]) Run(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	bm.Logger.Info("starting job manager", "queues", fmt.Sprintf("%v", bm.jobs))

	for queue := range bm.jobs {
		bm.Logger.Info("starting queue", "queue", queue)
		go bm.process(ctx, queue)
	}

	<-ctx.Done()
	return nil
}

func (bm *JobManager[T]) process(ctx context.Context, queue string) {
	for ctx.Err() == nil {
		bm.ProcessQueue(ctx, queue)
	}
}

// ProcessQueue processes all jobs in the given queue until the queue is empty.
// This is intended for use in tests.
func (jm *JobManager[T]) ProcessQueue(ctx context.Context, queue string) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	queueCh := make(chan string, 1)

	go func() {
		for ctx.Err() == nil {
			jsonPayload, err := jm.storage.Dequeue(ctx, queue)
			if err != nil {
				jm.Logger.Error("failed to pop job", "queue", queue, "error", err)
				continue
			}
			queueCh <- jsonPayload
		}
	}()

loop:
	for {
		select {
		case <-ctx.Done():
			cancel()
			break loop
		case <-time.After(1 * time.Millisecond):
			cancel()
			break loop
		case jsonPayload := <-queueCh:
			jm.processJob(ctx, queue, jsonPayload)
		}
	}
}

func (jm *JobManager[T]) processJob(ctx context.Context, queue string, jobPayload string) {
	if ctx.Err() != nil {
		return
	}

	t := jm.jobs[queue]
	// Handle pointers
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	value := reflect.New(t)

	err := json.Unmarshal([]byte([]byte(jobPayload)), value.Interface())
	job := value.Interface().(Job[T])

	if err != nil {
		jm.Logger.Error("failed to decode job JSON", "queue", queue, "error", err, "payload", jobPayload)
		return
	}

	func() {
		defer func() {
			if r := recover(); r != nil {
				jm.Logger.Error("panic in job", "queue", queue, "error", fmt.Sprint(r))
			}
		}()
		job.PerformJob(jm.jobContext)
	}()
}

// ProcessAll processes all jobs in all queues. This is useful in the test
// environment when you need to test behavior that includes job processing.
func (bm *JobManager[T]) ProcessAll(ctx context.Context) {
	for queue := range bm.jobs {
		bm.ProcessQueue(ctx, queue)
	}
}

// Enqueue enqueues a job using the given Storage passed to Manager.
// The queue name is automatically determined based on the type of Job as
// defined by RegisterQueue.
func (bm *JobManager[T]) Enqueue(ctx context.Context, job Job[T]) error {
	t := normalizeType(job)

	queueName, ok := bm.queueMap[t]
	if !ok {
		return fmt.Errorf("failed to find queue for %s", t.Name())
	}

	encoded, err := json.Marshal(job)

	if err != nil {
		return fmt.Errorf("failed to encode job: %v", err)
	}

	err = bm.storage.Enqueue(ctx, queueName, string(encoded))

	if err != nil {
		return fmt.Errorf("failed to push job to storage: %v", err)
	}

	return nil
}

func normalizeType(value any) reflect.Type {
	t := reflect.TypeOf(value)

	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	return t
}
