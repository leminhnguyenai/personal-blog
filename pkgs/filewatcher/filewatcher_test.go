package filewatcher_test

import (
	"context"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/leminhohoho/personal-blog/pkgs/filewatcher"
)

func runMultipleCommands(cmds [][]string) error {
	var err error
	for _, cmd := range cmds {
		name := cmd[0]
		args := cmd[1:]

		if err = exec.Command(name, args...).Run(); err != nil {
			return err
		}
	}

	return nil
}

func createTestFolder() error {
	var err error

	if err = runMultipleCommands([][]string{
		{"mkdir", "tmp"},
		{"mkdir", "tmp/folder"},
		{"touch", "tmp/a.txt"},
		{"touch", "tmp/b.txt"},
		{"touch", "tmp/folder/c.txt"},
	}); err != nil {
		return err
	}

	return nil
}

func testFileModification(fw *filewatcher.FileWatcher) func(t *testing.T) {
	return func(t *testing.T) {
		filesToModify := []string{"tmp/a.txt", "tmp/folder/c.txt"}

		for _, file := range filesToModify {
			ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*500)
			defer cancel()

			f, err := os.OpenFile(file, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
			if err != nil {
				t.Fatal(err)
			}
			defer f.Close()

			cmd := exec.Command("echo", "12345")
			cmd.Stdout = f

			go func() {
				time.Sleep(time.Millisecond * 100)
				if err := cmd.Run(); err != nil {
					t.Fatal(err)
				}
			}()

			select {
			case <-ctx.Done():
				t.Errorf("Timeout exceeded waiting for event from %s\n", file)
			case ev, ok := <-fw.Events:
				if !ok {
					t.Fatalf("Event channel closed enexpectedly\n")
				}

				if ev.Modified() {
					if ev.Path() != file {
						t.Errorf("Mismatch event path, want: %s, get: %s\n", file, ev.Path())
					}
				} else {
					t.Errorf("Not modification event at %s\n, event: %v", ev.Path(), ev)
				}
			}
		}
	}
}

func testFileCreation(fw *filewatcher.FileWatcher) func(t *testing.T) {
	return func(t *testing.T) {
		filesToCreate := []string{"tmp/d.txt", "tmp/folder/e.txt"}

		for _, file := range filesToCreate {
			ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*500)
			defer cancel()

			go func() {
				time.Sleep(time.Millisecond * 100)
				if err := exec.Command("touch", file).Run(); err != nil {
					t.Fatal(err)
				}
			}()

			select {
			case <-ctx.Done():
				t.Errorf("Timeout exceeded waiting for event from %s\n", file)
			case ev, ok := <-fw.Events:
				if !ok {
					t.Fatalf("Event channel closed enexpectedly\n")
				}

				if ev.Created() {
					if ev.Path() != file {
						t.Errorf("Mismatch event path, want: %s, get: %s\n", file, ev.Path())
					}
				} else {
					t.Errorf("Not modified event at %s\n", ev.Path())
				}
			}
		}
	}
}

func testFileDeletion(fw *filewatcher.FileWatcher) func(t *testing.T) {
	return func(t *testing.T) {
		filesToCreate := []string{"tmp/d.txt", "tmp/folder/e.txt"}

		for _, file := range filesToCreate {
			ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*500)
			defer cancel()

			go func() {
				time.Sleep(time.Millisecond * 100)
				if err := exec.Command("rm", file).Run(); err != nil {
					t.Fatal(err)
				}
			}()

			select {
			case <-ctx.Done():
				t.Errorf("Timeout exceeded waiting for event from %s\n", file)
			case ev, ok := <-fw.Events:
				if !ok {
					t.Fatalf("Event channel closed enexpectedly\n")
				}

				if ev.Deleted() {
					if ev.Path() != file {
						t.Errorf("Mismatch event path, want: %s, get: %s\n", file, ev.Path())
					}
				} else {
					t.Errorf("Not deletion event at %s\n", ev.Path())
				}
			}
		}
	}
}

func TestFileWatcher(t *testing.T) {
	fw := filewatcher.NewFileWatcher("tmp", time.Millisecond*100)
	defer fw.Close()
	time.Sleep(time.Millisecond * 500)

	if err := createTestFolder(); err != nil {
		t.Fatal(err)
	}

	t.Run("Test file modification", testFileModification(fw))
	t.Run("Test file creation", testFileCreation(fw))
	t.Run("Test file deletion", testFileDeletion(fw))
	t.Run("Test file modification", testFileModification(fw))

	if err := exec.Command("rm", "-rf", "tmp").Run(); err != nil {
		t.Fatal(err)
	}
}
