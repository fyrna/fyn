package task

import (
	"context"
	"fmt"
	"os"
	"slices"
	"strings"

	"github.com/fyrna/paws"
	"github.com/fyrna/x/color"
)

type CLI struct {
	runner *Runner
	parser *paws.Parser
}

func NewCLI(runner *Runner) *CLI {
	cli := &CLI{
		runner: runner,
		parser: paws.New(),
	}
	cli.setupParser()
	return cli
}

func (c *CLI) setupParser() {
	c.parser.
		PawBool("help", "h").END().
		PawBool("list", "l").END().
		PawBool("verbose", "v").END()

	c.parser.AddCommand([]string{"x"}, []*paws.Flag{
		{
			Name:    "parallel",
			Type:    paws.BoolType,
			Aliases: []string{"p"},
		},
		{
			Name:    "series",
			Type:    paws.BoolType,
			Aliases: []string{"s"},
		},
	})
}

func (c *CLI) Run(args []string) {
	if len(args) == 0 {
		c.printBanner()
		return
	}

	res, err := c.parser.Parse(args)
	if err != nil {
		fmt.Println(color.Wrap(color.Red, "Parse error nyaa~ 💫"))
		fmt.Println("Error:", err)
		os.Exit(1)
	}

	if res.Bool("help") {
		c.PrintHelp()
		return
	}

	if res.Bool("list") {
		c.PrintList()
		return
	}

	switch {
	case res.Command != nil && slices.Equal(res.Command.Path, []string{"x"}):
		c.handleRun(res)
	default:
		c.handleTask(args[0], res)
	}
}

func (c *CLI) handleRun(res *paws.ParseResult) {
	if len(res.Positional) == 0 {
		fmt.Println(color.Wrap(color.Yellow, "No tasks specified for run command~"))
		c.PrintHelp()
		return
	}

	tasks := res.Positional
	ctx := context.Background()

	var err error
	if res.Bool("parallel") {
		err = c.runner.Parallel(tasks...)(ctx)
	} else if res.Bool("series") {
		err = c.runner.Series(tasks...)(ctx)
	} else {
		// Default: run first task with dependencies
		err = c.runner.Run(ctx, tasks[0])
	}

	if err != nil {
		fmt.Println(color.Wrap(color.Red, "Task execution failed nyaaa >w< 💥"))
		fmt.Println("Error:", err)
		os.Exit(1)
	}

	fmt.Println(color.Wrap(color.BrightCyan, "All tasks completed nyan~ 🎉"))
}

func (c *CLI) handleTask(taskName string, res *paws.ParseResult) {
	if res.Bool("verbose") {
		fmt.Printf("Running task: %s\n", color.Wrap(color.BrightCyan, taskName))
	}

	if err := c.runner.Run(nil, taskName); err != nil {
		fmt.Println(color.Wrap(color.Red, "Task failed nyaaa >w< 💥"))
		fmt.Println("Error:", err)
		os.Exit(1)
	}

	fmt.Println(color.Wrap(color.BrightCyan, "Task finished nyan~ 🎉"))
}

func (c *CLI) printBanner() {
	fmt.Println(color.Wrap(color.BrightMagenta, "Task runner UwU 💕"))
	fmt.Println("a simple way to declare task-thing!")
	fmt.Println()
	fmt.Println(color.Wrap(color.BrightCyan+color.Bold, "Examples:"))
	fmt.Println("  task run build --parallel")
	fmt.Println("  task run deploy --series")
	fmt.Println("  task build --verbose")
	fmt.Println("  task --list")
	fmt.Println()
	fmt.Println(color.Wrap(color.Faint, "Use '--list' to see available tasks~ nya (ฅ^•ﻌ•^ฅ)"))
	fmt.Println(color.Wrap(color.Faint, "or use '--help' to print help message! nyan nyan!"))
}

func (c *CLI) PrintHelp() {
	fmt.Println(color.Wrap(color.Bold, "Very Cute Task Runner 💕"))
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  task [taskname]                    Run a single task")
	fmt.Println("  task run [tasks...] [flags]        Run multiple tasks")
	fmt.Println("  task --list                        List all tasks")
	fmt.Println("  task --help                        Show this help")
	fmt.Println()
	fmt.Println("Flags:")
	fmt.Println("  -h, --help                         Show help")
	fmt.Println("  -l, --list                         List tasks")
	fmt.Println("  -v, --verbose                      Verbose output")
	fmt.Println("  -p, --parallel                     Run tasks in parallel")
	fmt.Println("  -s, --series                       Run tasks in series")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  task build                         Run build task")
	fmt.Println("  task run build test --parallel     Run build and test in parallel")
	fmt.Println("  task run deploy --series           Run deploy tasks in series")
	fmt.Println("  task --verbose build               Run build with verbose output")
}

func (c *CLI) PrintList() {
	tasks := c.runner.ListTasks()

	if len(tasks) == 0 {
		fmt.Println(color.Wrap(color.Faint, "No tasks registered nya~ (´･ω･`)"))
		return
	}

	slices.SortFunc(tasks, func(a, b TaskInfo) int {
		return strings.Compare(a.Name, b.Name)
	})

	fmt.Println(color.Wrap(color.Bold, "Available tasks:"))
	for _, t := range tasks {
		var desc string
		if t.Desc != "" {
			desc = "- " + t.Desc
		}

		// Show dependencies if any
		var deps string
		if len(t.Deps) > 0 {
			deps = color.Wrap(color.Faint, fmt.Sprintf(" [deps: %s]", strings.Join(t.Deps, ", ")))
		}

		fmt.Printf("  %s%-15s%s %s%s\n",
			color.BrightCyan, t.Name, color.Reset, desc, deps)
	}
}
