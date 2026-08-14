package op

import (
	"context"
	"strconv"
	"testing"

	"github.com/bestruirui/octopus/internal/model"
)

func setTestSetting(t *testing.T, key model.SettingKey, value string) {
	t.Helper()
	previous, found := settingCache.Get(key)
	settingCache.Set(key, value)
	t.Cleanup(func() {
		if found {
			settingCache.Set(key, previous)
		} else {
			settingCache.Del(key)
		}
	})
}

func Test_RelayLogAdd_rebuilds_cache_backing_array_when_history_disabled(t *testing.T) {
	// Given
	setTestSetting(t, model.SettingKeyRelayLogKeepEnabled, strconv.FormatBool(false))
	relayLogCacheLock.Lock()
	relayLogCache = make([]model.RelayLog, relayLogMaxSizeNoDB-1, relayLogMaxSizeNoDB)
	for index := range relayLogCache {
		relayLogCache[index].ID = int64(index + 1)
	}
	relayLogCacheLock.Unlock()
	t.Cleanup(func() {
		relayLogCacheLock.Lock()
		relayLogCache = make([]model.RelayLog, 0, relayLogMaxSize)
		relayLogCacheLock.Unlock()
	})

	// When
	if err := RelayLogAdd(context.Background(), model.RelayLog{}); err != nil {
		t.Fatalf("RelayLogAdd() error = %v", err)
	}

	// Then
	relayLogCacheLock.Lock()
	defer relayLogCacheLock.Unlock()
	if len(relayLogCache) != relayLogMaxSizeNoDB/2 {
		t.Fatalf("cache length = %d, want %d", len(relayLogCache), relayLogMaxSizeNoDB/2)
	}
	if cap(relayLogCache) != relayLogMaxSizeNoDB {
		t.Fatalf("cache capacity = %d, want rebuilt capacity %d", cap(relayLogCache), relayLogMaxSizeNoDB)
	}
	if relayLogCache[0].ID != 51 {
		t.Fatalf("oldest retained log ID = %d, want 51", relayLogCache[0].ID)
	}
}

func Test_RelayLogSaveDBTask_rebuilds_cache_backing_array_when_history_disabled(t *testing.T) {
	// Given
	setTestSetting(t, model.SettingKeyRelayLogKeepEnabled, strconv.FormatBool(false))
	relayLogCacheLock.Lock()
	relayLogCache = make([]model.RelayLog, relayLogMaxSizeNoDB+1, relayLogMaxSizeNoDB+1)
	for index := range relayLogCache {
		relayLogCache[index].ID = int64(index + 1)
	}
	relayLogCacheLock.Unlock()
	t.Cleanup(func() {
		relayLogCacheLock.Lock()
		relayLogCache = make([]model.RelayLog, 0, relayLogMaxSize)
		relayLogCacheLock.Unlock()
	})

	// When
	if err := RelayLogSaveDBTask(context.Background()); err != nil {
		t.Fatalf("RelayLogSaveDBTask() error = %v", err)
	}

	// Then
	relayLogCacheLock.Lock()
	defer relayLogCacheLock.Unlock()
	if len(relayLogCache) != relayLogMaxSizeNoDB/2 {
		t.Fatalf("cache length = %d, want %d", len(relayLogCache), relayLogMaxSizeNoDB/2)
	}
	if cap(relayLogCache) != relayLogMaxSizeNoDB {
		t.Fatalf("cache capacity = %d, want rebuilt capacity %d", cap(relayLogCache), relayLogMaxSizeNoDB)
	}
	if relayLogCache[0].ID != 52 || relayLogCache[len(relayLogCache)-1].ID != 101 {
		t.Fatalf("retained log IDs = %d..%d, want 52..101", relayLogCache[0].ID, relayLogCache[len(relayLogCache)-1].ID)
	}
}
