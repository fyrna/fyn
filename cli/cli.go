package cli

import (
	"context"
	"fmt"
	"os"
	"slices"
	"strings"

	"github.com/fyrna/task"
)

type Options struct {
	DefaultTask string
	ShowListFn  func(r *task.Runner)
	ErrorFn     func(err error)
	Context     context.Context
}

type Option func(*Options)

func DefaultTask(name string) Option {
	return func(o *Options) { o.DefaultTask = name }
}

func ShowList(fn func(r *task.Runner)) Option {
	return func(o *Options) { o.ShowListFn = fn }
}

func ErrorHandler(fn func(error)) Option {
	return func(o *Options) { o.ErrorFn = fn }
}

func Context(ctx context.Context) Option {
	return func(o *Options) { o.Context = ctx }
}

func PrintHelp() {
	fmt.Println("task uwu runner :3")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  -h, --help     Show this help message")
	fmt.Println("  -l, --list     List all available tasks")
	fmt.Println()
	fmt.Println("if no task is specified, this help message will be shown.")
}

func Run(r *task.Runner, opts ...Option) {
	// defaults
	o := &Options{
		DefaultTask: "_",
		ShowListFn: func(r *task.Runner) {
			tasks := r.ListTasks()
			slices.SortFunc(tasks, func(a, b task.TaskInfo) int {
				return strings.Compare(a.Name, b.Name)
			})

			fmt.Println("Available tasks:")
			for _, t := range tasks {
				if t.Name == "_" {
					continue
				}
				desc := ""
				if t.Desc != "" {
					desc = t.Desc
				}
				fmt.Printf("  - %-15s %s\n", t.Name, desc)
			}
		},
		ErrorFn: func(err error) {
			fmt.Fprintln(os.Stderr, "Error:", err)
			os.Exit(1)
		},
		Context: context.Background(),
	}

	for _, opt := range opts {
		opt(o)
	}

	args := os.Args[1:]

	if len(args) == 0 {
		if err := r.Run(o.Context, o.DefaultTask); err != nil {
			if o.DefaultTask == "_" {
				PrintHelp()
				return
			}
			o.ErrorFn(err)
		}
		return
	}

	switch args[0] {
	case "--list", "-l":
		o.ShowListFn(r)
	case "--help", "-h":
		PrintHelp()
	default:
		if err := r.Run(o.Context, args[0]); err != nil {
			o.ErrorFn(err)
		}
	}
}
