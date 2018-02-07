package runner

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/howeyc/fsnotify"
	"time"
)

func watchFolder(path string) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		fatal(err)
	}

	go func() {
		for {
			select {
			case ev := <-watcher.Event:
				if isWatchedFile(ev.Name) {
					watcherLog("sending event %s", ev)
					startChannel <- ev.String()
				}
			case err := <-watcher.Error:
				watcherLog("error: %s", err)
			}
		}
	}()

	watcherLog("Watching %s", path)
	err = watcher.Watch(path)

	if err != nil {
		fatal(err)
	}
}

func watch() {
	root := root()
	filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() && !isTmpDir(path) {
			if len(path) > 1 && strings.HasPrefix(filepath.Base(path), ".") {
				return filepath.SkipDir
			}

			if isIgnoredFolder(path) {
				watcherLog("Ignoring %s", path)
				return filepath.SkipDir
			}

			watchFolder(path)
		}

		return err
	})
}

var size int64

func duWatcher() {
	root := root()
	go func(startChannel chan string, root string) {
		for {
			var currentSize int64
			filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
				if info.IsDir() && !isTmpDir(path) {
					if len(path) > 1 && strings.HasPrefix(filepath.Base(path), ".") {
						return filepath.SkipDir
					}

					if isIgnoredFolder(path) {
						watcherLog("Ignoring %s", path)
						return filepath.SkipDir
					}
				}
				if !info.IsDir() {
					if isWatchedFile(path) {
						currentSize += info.Size()
					}
				}

				return err
			})
			if currentSize != size {
				size = currentSize
				startChannel <- "ChangeSize"
			}
			time.Sleep(2 * time.Second)
		}
	}(startChannel, root)
}
