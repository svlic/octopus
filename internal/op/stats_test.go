package op

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/bestruirui/octopus/internal/db"
	"github.com/bestruirui/octopus/internal/model"
)

func Test_StatsSaveDB_restores_dirty_ids_when_persistence_fails(t *testing.T) {
	assertStatsSaveRestoresDirtyIDs(t, StatsSaveDB)
}

func Test_statsSaveDBWithDailyOverride_restores_dirty_ids_when_persistence_fails(t *testing.T) {
	assertStatsSaveRestoresDirtyIDs(t, func(ctx context.Context) error {
		return statsSaveDBWithDailyOverride(ctx, model.StatsDaily{})
	})
}

func assertStatsSaveRestoresDirtyIDs(t *testing.T, save func(context.Context) error) {
	t.Helper()

	// Given
	if err := db.InitDB("sqlite", filepath.Join(t.TempDir(), "octopus.db"), false); err != nil {
		t.Fatalf("InitDB() error = %v", err)
	}
	statsChannelCacheNeedUpdateLock.Lock()
	statsChannelCacheNeedUpdate = map[int]struct{}{11: {}}
	statsChannelCacheNeedUpdateLock.Unlock()
	statsModelCacheNeedUpdateLock.Lock()
	statsModelCacheNeedUpdate = map[int]struct{}{22: {}}
	statsModelCacheNeedUpdateLock.Unlock()
	statsAPIKeyCacheNeedUpdateLock.Lock()
	statsAPIKeyCacheNeedUpdate = map[int]struct{}{33: {}}
	statsAPIKeyCacheNeedUpdateLock.Unlock()
	t.Cleanup(func() {
		statsChannelCacheNeedUpdateLock.Lock()
		statsChannelCacheNeedUpdate = make(map[int]struct{})
		statsChannelCacheNeedUpdateLock.Unlock()
		statsModelCacheNeedUpdateLock.Lock()
		statsModelCacheNeedUpdate = make(map[int]struct{})
		statsModelCacheNeedUpdateLock.Unlock()
		statsAPIKeyCacheNeedUpdateLock.Lock()
		statsAPIKeyCacheNeedUpdate = make(map[int]struct{})
		statsAPIKeyCacheNeedUpdateLock.Unlock()
	})
	if err := db.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}

	// When
	err := save(context.Background())

	// Then
	if err == nil {
		t.Fatal("stats save error = nil, want persistence failure")
	}
	statsChannelCacheNeedUpdateLock.Lock()
	_, channelDirty := statsChannelCacheNeedUpdate[11]
	statsChannelCacheNeedUpdateLock.Unlock()
	statsModelCacheNeedUpdateLock.Lock()
	_, modelDirty := statsModelCacheNeedUpdate[22]
	statsModelCacheNeedUpdateLock.Unlock()
	statsAPIKeyCacheNeedUpdateLock.Lock()
	_, apiKeyDirty := statsAPIKeyCacheNeedUpdate[33]
	statsAPIKeyCacheNeedUpdateLock.Unlock()
	if !channelDirty || !modelDirty || !apiKeyDirty {
		t.Fatalf("dirty IDs restored = channel:%t model:%t apiKey:%t, want all true", channelDirty, modelDirty, apiKeyDirty)
	}
}
