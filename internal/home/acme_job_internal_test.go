package home

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAcmeJob_Replay(t *testing.T) {
	j := newAcmeJob()

	j.log("info", "line one")
	j.log("info", "line two")

	lines, _, done, result := j.snapshot()
	require.Len(t, lines, 2)
	assert.Equal(t, "line one", lines[0].Message)
	assert.Equal(t, "line two", lines[1].Message)
	assert.False(t, done)
	assert.Nil(t, result)
}

func TestAcmeJob_Finish(t *testing.T) {
	j := newAcmeJob()

	j.log("info", "starting")
	j.finish(&acmeJobResult{Success: true})

	lines, _, done, result := j.snapshot()
	require.Len(t, lines, 1)
	assert.True(t, done)
	require.NotNil(t, result)
	assert.True(t, result.Success)

	// Further log/finish calls after done must be no-ops.
	j.log("info", "should be ignored")
	j.finish(&acmeJobResult{Success: false})

	lines, _, done, result = j.snapshot()
	assert.Len(t, lines, 1)
	assert.True(t, done)
	assert.True(t, result.Success)
	assert.True(t, j.isDone())
}

func TestAcmeJob_NotifyWakesSubscriber(t *testing.T) {
	j := newAcmeJob()

	_, notify, _, _ := j.snapshot()

	var wg sync.WaitGroup
	wg.Add(1)

	woke := false
	go func() {
		defer wg.Done()

		select {
		case <-notify:
			woke = true
		case <-time.After(2 * time.Second):
		}
	}()

	j.log("info", "wake up")
	wg.Wait()

	assert.True(t, woke, "subscriber should have woken up on new log line")
}

func TestAcmeJob_NotifyWakesOnFinish(t *testing.T) {
	j := newAcmeJob()

	_, notify, _, _ := j.snapshot()

	var wg sync.WaitGroup
	wg.Add(1)

	woke := false
	go func() {
		defer wg.Done()

		select {
		case <-notify:
			woke = true
		case <-time.After(2 * time.Second):
		}
	}()

	j.finish(&acmeJobResult{Success: true})
	wg.Wait()

	assert.True(t, woke, "subscriber should have woken up on finish")
}

func TestAcmeJob_ConcurrentAccess(t *testing.T) {
	j := newAcmeJob()

	var wg sync.WaitGroup
	for range 20 {
		wg.Add(1)
		go func() {
			defer wg.Done()

			j.log("info", "concurrent line")
		}()
	}

	wg.Wait()
	j.finish(&acmeJobResult{Success: true})

	lines, _, done, result := j.snapshot()
	assert.Len(t, lines, 20)
	assert.True(t, done)
	assert.True(t, result.Success)
}
