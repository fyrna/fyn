package fytask

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
)

type TaskFn func() error

type Runner struct {
	tasks map[string]TaskFn
	mu    sync.Mutex
}

var defaultRunner = New()

func New() *Runner {
	return &Runner{
		tasks: make(map[string]TaskFn),
	}
}

func (r *Runner) Task(name string, fn TaskFn) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.tasks[name] = fn
}

func (r *Runner) Run(name string) {
	task, ok := r.tasks[name]
	if !ok {
		panic(fmt.Sprintf("task not found: %s", name))
	}
	if err := task(); err != nil {
		panic(fmt.Sprintf("task %s failed: %v", name, err))
	}
}

func Task(name string, fn TaskFn) {
	defaultRunner.Task(name, fn)
}

func Run(name string) {
	defaultRunner.Run(name)
}

func Series(names ...string) TaskFn {
	return func() error {
		for _, n := range names {
			defaultRunner.Run(n)
		}
		return nil
	}
}

func Parallel(names ...string) TaskFn {
	return func() error {
		var wg sync.WaitGroup
		errs := make(chan error, len(names))

		for _, n := range names {
			wg.Add(1)

			go func(taskName string) {
				defer wg.Done()
				task, ok := defaultRunner.tasks[taskName]
				if !ok {
					errs <- fmt.Errorf("task not found: %s", taskName)
					return
				}
				if err := task(); err != nil {
					errs <- fmt.Errorf("task %s failed: %v", taskName, err)
				}
			}(n)
		}

		wg.Wait()
		close(errs)

		if len(errs) > 0 {
			return <-errs
		}
		return nil
	}
}

// no output
func Sh(cmd string, args ...string) error {
	c := exec.Command(cmd, args...)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	return c.Run()
}

// returns stdout as string.
func ShOut(cmd string, args ...string) string {
	var out bytes.Buffer

	c := exec.Command(cmd, args...)
	c.Stdout = &out
	c.Stderr = os.Stderr

	if err := c.Run(); err != nil {
		panic(err)
	}
	return out.String()
}

func MustSh(cmd string, args ...string) {
	if err := Sh(cmd, args...); err != nil {
		panic(err)
	}
}

func Log(v ...any) {
	fmt.Println(v...)
}

func Logf(format string, v ...any) {
	fmt.Printf(format+"\n", v...)
}

func Copy(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}

func Rm(path string) error {
	return os.RemoveAll(path)
}

func Mkdir(path string) error {
	return os.MkdirAll(path, 0755)
}

func Env(key string) string {
	return os.Getenv(key)
}

func SetEnv(key, val string) {
	os.Setenv(key, val)
}
