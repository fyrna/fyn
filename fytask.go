package fytask

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
)

type TaskFn func(ctx context.Context) error

type Runner struct {
	tasks map[string]TaskFn
	mu    sync.Mutex
}

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

func (r *Runner) Run(ctx context.Context, name string) error {
	task, ok := r.tasks[name]
	if !ok {
		return fmt.Errorf("task '%s' not found", name)
	}
	return task(ctx)
}

// run with context.Background
func (r *Runner) Rawr(name string) error {
	return r.Run(context.Background(), name)
}

func (r *Runner) List() {
	fmt.Println("Available tasks:")
	for name := range r.tasks {
		fmt.Println("  ", name)
	}
}

func Series(tasks ...string) TaskFn {
	return func(ctx context.Context) error {
		for _, t := range tasks {
			if err := defaultRunner.Run(ctx, t); err != nil {
				return err
			}
		}
		return nil
	}
}

func Parallel(tasks ...string) TaskFn {
	return func(ctx context.Context) error {
		var wg sync.WaitGroup
		errs := make(chan error, 1)

		for _, t := range tasks {
			wg.Add(1)
			go func(taskName string) {
				defer wg.Done()
				if err := defaultRunner.Run(ctx, taskName); err != nil {
					select {
					case errs <- err:
					default:
					}
				}
			}(t)
		}

		wg.Wait()
		close(errs)
		return <-errs
	}
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

func If(condition bool, task string) TaskFn {
	return func(ctx context.Context) error {
		if condition {
			return defaultRunner.Run(ctx, task)
		}
		return nil
	}
}

func Unless(condition bool, task string) TaskFn {
	return If(!condition, task)
}

func Exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

var defaultRunner = New()

func Task(name string, fn TaskFn) {
	defaultRunner.Task(name, fn)
}

func Run(ctx context.Context, name string) error {
	return defaultRunner.Run(ctx, name)
}
