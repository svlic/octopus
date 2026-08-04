package task

import (
	"testing"
	"time"
)

func Test_Update_keeps_latest_interval_before_runner_starts(t *testing.T) {
	// Given
	resetTaskRegistry(t)
	Register("test", time.Hour, false, func() {})

	// When
	Update("test", 2*time.Hour)

	// Then
	entry := registeredTask(t, "test")
	select {
	case interval := <-entry.updateCh:
		if interval != 2*time.Hour {
			t.Fatalf("queued interval = %v, want 2h", interval)
		}
	default:
		t.Fatal("task update was dropped before runner started")
	}
}

func Test_Update_reenables_paused_task(t *testing.T) {
	// Given
	resetTaskRegistry(t)
	runs := make(chan struct{}, 1)
	Register("test", 0, false, func() {
		runs <- struct{}{}
	})
	entry := registeredTask(t, "test")
	go runTask(entry)
	t.Cleanup(func() {
		close(entry.stopCh)
	})

	// When
	Update("test", 5*time.Millisecond)

	// Then
	select {
	case <-runs:
	case <-time.After(time.Second):
		t.Fatal("paused task did not run after a positive interval update")
	}
}

func resetTaskRegistry(t *testing.T) {
	t.Helper()
	tasksMu.Lock()
	previous := tasks
	tasks = make(map[string]*taskEntry)
	tasksMu.Unlock()
	t.Cleanup(func() {
		tasksMu.Lock()
		tasks = previous
		tasksMu.Unlock()
	})
}

func registeredTask(t *testing.T, name string) *taskEntry {
	t.Helper()
	tasksMu.RLock()
	defer tasksMu.RUnlock()
	entry, exists := tasks[name]
	if !exists {
		t.Fatalf("task %q is not registered", name)
	}
	return entry
}
