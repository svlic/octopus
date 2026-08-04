package task

import (
	"sync"
	"time"

	"github.com/bestruirui/octopus/internal/utils/log"
)

type taskEntry struct {
	name       string
	interval   time.Duration
	fn         func()
	runOnStart bool
	ticker     *time.Ticker
	stopCh     chan struct{}
	updateCh   chan time.Duration
}

var (
	tasks   = make(map[string]*taskEntry)
	tasksMu sync.RWMutex
)

// Register 注册一个定时任务
// runOnStart: 是否在启动时立即执行一次
func Register(name string, interval time.Duration, runOnStart bool, fn func()) {
	tasksMu.Lock()
	defer tasksMu.Unlock()

	if _, exists := tasks[name]; exists {
		log.Warnf("task %s already registered, skipping", name)
		return
	}

	tasks[name] = &taskEntry{
		name:       name,
		interval:   interval,
		fn:         fn,
		runOnStart: runOnStart,
		stopCh:     make(chan struct{}),
		updateCh:   make(chan time.Duration, 1),
	}
	log.Debugf("task %s registered with interval %v, runOnStart: %v", name, interval, runOnStart)
}

// Update 更新任务的执行间隔；非正值暂停任务，之后可通过正值恢复。
func Update(name string, interval time.Duration) {
	tasksMu.Lock()
	entry, exists := tasks[name]
	if !exists {
		tasksMu.Unlock()
		log.Warnf("task %s not found", name)
		return
	}
	select {
	case <-entry.updateCh:
	default:
	}
	entry.updateCh <- interval
	tasksMu.Unlock()

	log.Infof("task %s interval updated to %v", name, interval)
}

// RUN 启动所有注册的任务
func RUN() {
	tasksMu.RLock()
	for _, entry := range tasks {
		go runTask(entry)
	}
	tasksMu.RUnlock()

	// 阻塞主协程
	select {}
}

func runTask(entry *taskEntry) {
	var tickerC <-chan time.Time
	if entry.interval > 0 {
		entry.ticker = time.NewTicker(entry.interval)
		tickerC = entry.ticker.C
		if entry.runOnStart {
			go entry.fn()
		}
	}

	for {
		select {
		case <-tickerC:
			go entry.fn()
		case newInterval := <-entry.updateCh:
			if entry.ticker != nil {
				entry.ticker.Stop()
				entry.ticker = nil
				tickerC = nil
			}
			if newInterval > 0 {
				entry.ticker = time.NewTicker(newInterval)
				tickerC = entry.ticker.C
			}
		case <-entry.stopCh:
			if entry.ticker != nil {
				entry.ticker.Stop()
			}
			return
		}
	}
}
