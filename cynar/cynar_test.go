package cynar

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

var run bool = false

type fakeJob struct {
	Value string
}

func (f fakeJob) PerformJob(jobContext string) {
	if jobContext != "jobContext" {
		panic("invalid context")
	}

	if f.Value != "omg" {
		panic("invalid value")
	}
	run = true
}

var jobContext string = "jobContext"

func TestEnqueueAndPerform(t *testing.T) {
	run = false
	bm := New(NewMemoryStorage(), jobContext)
	bm.RegisterQueue("test", &fakeJob{})
	err := bm.PushJob(context.TODO(), &fakeJob{Value: "omg"})

	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	require.False(t, run)

	go bm.Run(ctx)

	time.Sleep(10 * time.Millisecond)
	cancel()

	require.True(t, run)
}

func TestEnqueueAndPerformNonPointer(t *testing.T) {
	run = false
	bm := New(NewMemoryStorage(), jobContext)
	bm.RegisterQueue("test", fakeJob{})
	err := bm.PushJob(context.TODO(), fakeJob{Value: "omg"})

	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	require.False(t, run)

	go bm.Run(ctx)

	time.Sleep(10 * time.Millisecond)
	cancel()

	require.True(t, run)
}

func TestHandleEmptyQueue(t *testing.T) {
	run = false
	bm := New(NewMemoryStorage(), jobContext)
	bm.RegisterQueue("test", fakeJob{})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	require.False(t, run)

	go bm.Run(ctx)
	require.False(t, run)

	require.False(t, run)
}

func TestPushJob_NoQueueDefined(t *testing.T) {
	bm := New(NewMemoryStorage(), jobContext)
	err := bm.PushJob(context.TODO(), &fakeJob{})

	require.ErrorContains(t, err, "failed to find queue for fakeJob")
}

func ExampleManager_RegisterQueue() {
	bm := New(NewMemoryStorage(), jobContext)

	bm.RegisterQueue("test", &fakeJob{})
}
