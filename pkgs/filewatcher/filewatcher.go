package filewatcher

import (
	"errors"
	"os"
	"path"
	"slices"
	"time"
)

type EventType int

const (
	Creation EventType = iota
	Modification
	Deletion
)

// The event which is sent to the channel when file changes detected
type Event struct {
	path      string
	eventType EventType
}

// Check if the event is of modification, return true if it is a modification event
func (e *Event) Modified() bool {
	return e.eventType == Modification
}

// Check if the event is of creation, return true if it is a creation event
func (e *Event) Created() bool {
	return e.eventType == Creation
}

// Check if the event is of deletion, return true if it is a deletion event
func (e *Event) Deleted() bool {
	return e.eventType == Deletion
}

// Return the path of the file that trigger the event
func (e *Event) Path() string {
	return e.path
}

type FileToWatch struct {
	Path    string
	modTime time.Time
}

type FileWatcher struct {
	Events chan Event
	Errors chan error

	// firstScan is for indication for the first scan, will be set to false afterward
	firstScan   bool
	initialized chan bool
	interval    time.Duration
	// The dir to be watched, only files within this dir will be watched
	rootDir string
	// List of file paths those are current watched
	watchList []*FileToWatch
}

// Return a file watcher and start listening to the directory specified
func NewFileWatcher(dirPath string, interval time.Duration) *FileWatcher {
	fw := &FileWatcher{
		Events:      make(chan Event, 100),
		Errors:      make(chan error, 100),
		firstScan:   true,
		initialized: make(chan bool),
		interval:    interval,
		rootDir:     dirPath,
	}

	go fw.run()

	return fw
}

// Close all the channel of the file watcher
func (fw *FileWatcher) Close() {
	close(fw.Events)
	close(fw.Errors)
	close(fw.initialized)
}

func (fw *FileWatcher) run() {
	ticker := time.NewTicker(fw.interval)

	for {
		select {
		case <-ticker.C:
			if err := fw.updateWatchList(); err != nil {
				fw.Errors <- err
			}

			if err := fw.scanWatchList(); err != nil {
				fw.Errors <- err
			}
		}
	}
}

func (fw *FileWatcher) updateWatchList() error {
	dirList := []string{fw.rootDir}

	for len(dirList) > 0 {
		var newDirList []string
		for _, dir := range dirList {
			entries, err := os.ReadDir(dir)
			if err != nil {
				return err
			}

			for _, entry := range entries {
				if entry.IsDir() {
					newDirList = append(newDirList, path.Join(dir, entry.Name()))
					continue
				}
				newFile := path.Join(dir, entry.Name())

				if !slices.ContainsFunc(fw.watchList, func(f *FileToWatch) bool {
					return f.Path == newFile
				}) {
					fw.watchList = append(fw.watchList, &FileToWatch{newFile, time.Time{}})
					if fw.firstScan {
						continue
					}
					fw.Events <- Event{newFile, Creation}
				}
			}
		}

		dirList = newDirList
	}

	if fw.firstScan {
		fw.firstScan = false
		fw.initialized <- true
	}

	return nil
}

func (fw *FileWatcher) scanWatchList() error {
	var newWatchList []*FileToWatch

	for _, file := range fw.watchList {
		fileInfo, err := os.Stat(file.Path)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				fw.Events <- Event{file.Path, Deletion}
				continue
			}

			return err
		}

		newWatchList = append(newWatchList, file)

		if file.modTime.IsZero() {
			file.modTime = fileInfo.ModTime()
			continue
		}

		if !file.modTime.Equal(fileInfo.ModTime()) {
			fw.Events <- Event{file.Path, Modification}
			file.modTime = fileInfo.ModTime()
			continue
		}
	}

	fw.watchList = newWatchList

	return nil
}

func (fw *FileWatcher) GetWatchList() []FileToWatch {
	var watchList []FileToWatch

	if fw.firstScan {
		select {
		case <-fw.initialized:
			goto RETURN_WATCHLIST
		}
	}
	goto RETURN_WATCHLIST

RETURN_WATCHLIST:
	for _, file := range fw.watchList {
		watchList = append(watchList, *file)
	}
	return watchList
}
