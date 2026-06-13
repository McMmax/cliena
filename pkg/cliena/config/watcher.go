package config

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/pelletier/go-toml/v2"
)

// debounceWindow coalesces the burst of write/create events editors emit when
// saving a file, so one save triggers exactly one reload.
const debounceWindow = 200 * time.Millisecond

// ConfigReloadCallback receives the reparsed config, or an error on failure.
type ConfigReloadCallback func(newConfig *UserConfig, err error)

// Watcher reloads the config file when it changes on disk.
type Watcher struct {
	fsw      *fsnotify.Watcher
	path     string
	callback ConfigReloadCallback
	cancel   context.CancelFunc
	done     chan struct{} // closed once the watch loop has fully exited
	stopOnce sync.Once
}

// NewWatcher starts watching configPath, invoking callback on each change.
func NewWatcher(configPath string, callback ConfigReloadCallback) (*Watcher, error) {
	fsw, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	if err := fsw.Add(configPath); err != nil {
		_ = fsw.Close()
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())
	w := &Watcher{
		fsw:      fsw,
		path:     configPath,
		callback: callback,
		cancel:   cancel,
		done:     make(chan struct{}),
	}

	go w.run(ctx)
	return w, nil
}

// run is the watch loop: it debounces write/create bursts, logs watcher errors,
// and exits cleanly on context cancel.
func (w *Watcher) run(ctx context.Context) {
	defer close(w.done)

	var timer *time.Timer
	defer func() {
		if timer != nil {
			timer.Stop()
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case event, ok := <-w.fsw.Events:
			if !ok {
				return
			}
			if event.Has(fsnotify.Write) || event.Has(fsnotify.Create) {
				if timer != nil {
					timer.Stop()
				}
				timer = time.AfterFunc(debounceWindow, w.reload)
			}
		case err, ok := <-w.fsw.Errors:
			if !ok {
				return
			}
			log.Printf("config watcher error: %v", err)
		}
	}
}

// reload reparses the file and forwards the result to the callback.
func (w *Watcher) reload() {
	cfg, err := ReloadConfig(w.path)
	w.callback(cfg, err) // cfg is nil on error, matching the callback contract
}

// Stop cancels the watch loop and releases the fsnotify handle. Idempotent.
func (w *Watcher) Stop() {
	w.stopOnce.Do(func() {
		w.cancel()
		_ = w.fsw.Close()
		<-w.done // wait for the loop to drain before returning
	})
}

// ReloadConfig reads, parses, backfills, and validates the config at path.
// Validation errors are fatal; warnings are dropped on this path.
func ReloadConfig(path string) (*UserConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg UserConfig
	if err := toml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	applyDefaults(&cfg)

	if v := ValidateConfig(&cfg); v.HasErrors() {
		return nil, fmt.Errorf("config has %d error(s)", len(v.Errors))
	}
	return &cfg, nil
}
