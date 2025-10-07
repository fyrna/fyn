package cli

import (
	"flag"
	"fmt"
	"os"
	"slices"
	"strings"

	"github.com/fyrna/x/color"

	"github.com/fyrna/fn/task"
)

var (
	showHelp = flag.Bool("help", false, "show help")
	showList = flag.Bool("list", false, "list all tasks")
)

func Run(t *task.Runner) {
	flag.Parse()

	args := flag.Args()

	if len(args) == 0 && !*showList && !*showHelp {
		fmt.Println(color.Wrap(color.BrightMagenta, "Task runner UwU 💕"))
		fmt.Println("a simple way to declare task-thing!")
		fmt.Println()
		fmt.Println(color.Wrap(color.Faint, "Use '--list' to see available tasks~ nya (ฅ^•ﻌ•^ฅ)"))
		fmt.Println(color.Wrap(color.Faint, "or use '--help' to print help message! nyan nyan!"))
		return
	}

	if *showHelp {
		PrintHelp()
		return
	}

	if *showList {
		PrintList(t)
		return
	}

	taskName := args[0]

	if err := t.Run(nil, taskName); err != nil {
		fmt.Println(color.Wrap(color.Red, "Task failed nyaaa >w< 💥"))
		fmt.Println("Error:", err)
		os.Exit(1)
	}

	fmt.Println(color.Wrap(color.BrightCyan, "Task finished nyan~ 🎉"))
}

func PrintHelp() {
	fmt.Println(color.Wrap(color.Bold, "Very Cute Task Runner 💕"))
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  task [taskname]      Run a task")
	fmt.Println("  task --list          List all tasks")
	fmt.Println("  task --help          Show this help")
	fmt.Println()
}

func PrintList(t *task.Runner) {
	tasks := t.ListTasks()

	if len(tasks) == 0 {
		fmt.Println(color.Wrap(color.Faint, "No tasks registered nya~ (´･ω･`)"))
		return
	}

	slices.SortFunc(tasks, func(a, b task.TaskInfo) int {
		return strings.Compare(a.Name, b.Name)
	})

	fmt.Println(color.Wrap(color.Bold, "Available tasks:"))
	for _, t := range tasks {

		var desc string
		if t.Desc != "" {
			desc = "- " + t.Desc
		}
		fmt.Printf("  %s%-15s%s %s\n",
			color.BrightCyan, t.Name, color.Reset, desc)
	}
}
