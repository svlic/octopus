package op

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/bestruirui/octopus/internal/db"
	"github.com/bestruirui/octopus/internal/model"
)

func Test_RelayLogClear_waits_for_in_flight_flush(t *testing.T) {
	// Given
	if err := db.InitDB("sqlite", filepath.Join(t.TempDir(), "octopus.db"), false); err != nil {
		t.Fatalf("InitDB() error = %v", err)
	}
	t.Cleanup(func() {
		_ = db.Close()
	})

	relayLogFlushLock.Lock()
	clearResult := make(chan error, 1)
	go func() {
		clearResult <- RelayLogClear(context.Background())
	}()

	// When
	select {
	case err := <-clearResult:
		relayLogFlushLock.Unlock()
		t.Fatalf("RelayLogClear() returned before active flush completed: %v", err)
	case <-time.After(100 * time.Millisecond):
		relayLogFlushLock.Unlock()
	}

	// Then
	select {
	case err := <-clearResult:
		if err != nil {
			t.Fatalf("RelayLogClear() error = %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("RelayLogClear() did not finish after active flush completed")
	}
}

func Test_relayLogCleanup_waits_for_in_flight_flush(t *testing.T) {
	// Given
	if err := db.InitDB("sqlite", filepath.Join(t.TempDir(), "octopus.db"), false); err != nil {
		t.Fatalf("InitDB() error = %v", err)
	}
	setTestSetting(t, model.SettingKeyRelayLogKeepPeriod, "30")
	t.Cleanup(func() {
		_ = db.Close()
	})

	relayLogFlushLock.Lock()
	cleanupResult := make(chan error, 1)
	go func() {
		cleanupResult <- relayLogCleanup(context.Background())
	}()

	// When
	select {
	case err := <-cleanupResult:
		relayLogFlushLock.Unlock()
		t.Fatalf("relayLogCleanup() returned before active flush completed: %v", err)
	case <-time.After(100 * time.Millisecond):
		relayLogFlushLock.Unlock()
	}

	// Then
	select {
	case err := <-cleanupResult:
		if err != nil {
			t.Fatalf("relayLogCleanup() error = %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("relayLogCleanup() did not finish after active flush completed")
	}
}

func Test_RelayLogClear_removes_database_and_cached_logs(t *testing.T) {
	// Given
	if err := db.InitDB("sqlite", filepath.Join(t.TempDir(), "octopus.db"), false); err != nil {
		t.Fatalf("InitDB() error = %v", err)
	}
	t.Cleanup(func() {
		_ = db.Close()
	})
	if err := db.GetDB().Create(&model.RelayLog{ID: 1, Time: time.Now().Unix()}).Error; err != nil {
		t.Fatalf("create relay log: %v", err)
	}
	relayLogCacheLock.Lock()
	relayLogCache = []model.RelayLog{{ID: 2, Time: time.Now().Unix()}}
	relayLogCacheLock.Unlock()
	t.Cleanup(func() {
		relayLogCacheLock.Lock()
		relayLogCache = make([]model.RelayLog, 0, relayLogMaxSize)
		relayLogCacheLock.Unlock()
	})

	// When
	if err := RelayLogClear(context.Background()); err != nil {
		t.Fatalf("RelayLogClear() error = %v", err)
	}

	// Then
	var count int64
	if err := db.GetDB().Model(&model.RelayLog{}).Count(&count).Error; err != nil {
		t.Fatalf("count relay logs: %v", err)
	}
	relayLogCacheLock.Lock()
	cacheCount := len(relayLogCache)
	relayLogCacheLock.Unlock()
	if count != 0 || cacheCount != 0 {
		t.Fatalf("logs after clear = database:%d cache:%d, want both 0", count, cacheCount)
	}
}
