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

type Event struct {
	path      string
	eventType EventType
}

func (e *Event) Modified() bool {
	return e.eventType == Modification
}

func (e *Event) Created() bool {
	return e.eventType == Creation
}

func (e *Event) Deleted() bool {
	return e.eventType == Deletion
}

func (e *Event) Path() string {
	return e.path
}

type fileToWatch struct {
	path    string
	modTime time.Time
}

type FileWatcher struct {
	Events chan Event
	Errors chan error

	// firstScan is for indication for the first scan, will be set to false afterward
	firstScan bool
	interval  time.Duration
	// The dir to be watched, only files within this dir will be watched
	rootDir string
	// List of file paths those are current watched
	watchList []*fileToWatch
}

func NewFileWatcher(dirPath string, interval time.Duration) *FileWatcher {
	fw := &FileWatcher{
		Events:    make(chan Event, 100),
		Errors:    make(chan error, 100),
		firstScan: true,
		interval:  interval,
		rootDir:   dirPath,
	}

	go fw.run()

	return fw
}

func (fw *FileWatcher) Close() {
	close(fw.Events)
	close(fw.Errors)
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

				if !slices.ContainsFunc(fw.watchList, func(f *fileToWatch) bool {
					return f.path == newFile
				}) {
					fw.watchList = append(fw.watchList, &fileToWatch{newFile, time.Time{}})
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
	}

	return nil
}

func (fw *FileWatcher) scanWatchList() error {
	var newWatchList []*fileToWatch

	for _, file := range fw.watchList {
		fileInfo, err := os.Stat(file.path)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				fw.Events <- Event{file.path, Deletion}
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
			fw.Events <- Event{file.path, Modification}
			file.modTime = fileInfo.ModTime()
			continue
		}
	}

	fw.watchList = newWatchList

	return nil
}
