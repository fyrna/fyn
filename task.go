package task

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
	"sync"
)

var (
	ErrTaskNotFound       = errors.New("task not found")
	ErrCircularDependency = errors.New("circular dependency detected")
)

type Errors []error

func (e Errors) Error() string {
	if len(e) == 0 {
		return ""
	}
	if len(e) == 1 {
		return e[0].Error()
	}

	var sb strings.Builder
	sb.WriteString("errors:\n")

	for i, err := range e {
		sb.WriteString(fmt.Sprintf(" %d. %s\n", i+1, err))
	}

	return sb.String()
}

func (e Errors) Unwrap() []error { return []error(e) }

type TaskFn func(ctx context.Context) error

type TaskInfo struct {
	Name string
	Desc string
	Deps []string
	Func TaskFn
}

type Runner struct {
	mu    sync.Mutex
	tasks map[string]*TaskInfo
}

func New() *Runner {
	return &Runner{
		tasks: make(map[string]*TaskInfo),
	}
}

func (r *Runner) Unit(name string, fn TaskFn) {
	r.AddUnit(name, "", nil, fn)
}

func (r *Runner) AddUnit(name, desc string, deps []string, fn TaskFn) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.tasks[name] = &TaskInfo{
		Name: name,
		Desc: desc,
		Deps: deps,
		Func: fn,
	}
}

func (r *Runner) Validate() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	var errs Errors

	// check circular deps
	for t := range r.tasks {
		if err := r.detectCircular(t, nil); err != nil {
			errs = append(errs, err)
		}
	}

	// check missing deps
	for name, info := range r.tasks {
		for _, dep := range info.Deps {
			if _, ok := r.tasks[dep]; !ok {
				errs = append(errs, fmt.Errorf("task '%s': dependency '%s' not found", name, dep))
			}
		}
	}

	if len(errs) > 0 {
		return errs
	}

	return nil
}

func (r *Runner) Run(ctx context.Context, name string) error {
	r.mu.Lock()
	task, ok := r.tasks[name]
	r.mu.Unlock()

	if ctx == nil {
		ctx = context.Background()
	}

	if !ok {
		return fmt.Errorf("%w: '%s'", ErrTaskNotFound, name)
	}

	// run deps first
	for _, dep := range task.Deps {
		if err := r.Run(ctx, dep); err != nil {
			return fmt.Errorf("dependency '%s' failed: %w", dep, err)
		}
	}

	return task.Func(ctx)
}

func (r *Runner) ListTasks() []TaskInfo {
	r.mu.Lock()
	defer r.mu.Unlock()

	res := make([]TaskInfo, 0, len(r.tasks))

	for _, info := range r.tasks {
		res = append(res, *info)
	}

	return res
}

func (r *Runner) detectCircular(current string, path []string) error {
	if slices.Contains(path, current) {
		return fmt.Errorf("%w: %s", ErrCircularDependency,
			strings.Join(append(path, current), " -> "))
	}

	task, ok := r.tasks[current]
	if !ok {
		return nil
	}

	newPath := append(path, current)
	for _, dep := range task.Deps {
		if err := r.detectCircular(dep, newPath); err != nil {
			return err
		}
	}

	return nil
}

func (r *Runner) Series(tasks ...string) TaskFn {
	return func(ctx context.Context) error {
		for _, t := range tasks {
			if err := r.Run(ctx, t); err != nil {
				return err
			}
		}
		return nil
	}
}

func (r *Runner) Parallel(tasks ...string) TaskFn {
	return func(ctx context.Context) error {
		var wg sync.WaitGroup
		var mu sync.Mutex
		var errs Errors

		for _, t := range tasks {
			wg.Add(1)
			go func(taskName string) {
				defer wg.Done()
				if err := r.Run(ctx, taskName); err != nil {
					mu.Lock()
					errs = append(errs, fmt.Errorf("task %s: %w", taskName, err))
					mu.Unlock()
				}
			}(t)
		}

		wg.Wait()

		if len(errs) > 0 {
			return errs
		}

		return nil
	}
}
