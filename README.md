# fyrna/task

i want to build my own task runner, but i realize it fking hard and **WORTHLESS** to designing new domain spesific language JUST FOR **STUPID PROJECT**

so, i think, can i create my own inside Go?

### quick example

```go
//go:build ignore
package main

import (
	"context"

	"github.com/fyrna/task"
	S "github.com/fyrna/task/shell"
)

func main() {
	t := task.New()

	t.Task("file", func(ctx context.Context) error {
		S.S(`echo "miaw" > file.txt`)
		S.S("cat file.txt")
		S.S("rm -f file.txt")

		S.Exec(ctx,
			"for i in (seq 1 3); echo $i; end",
			&S.Options{
				Shell: "fish",
			})

		return nil
	})

	t.Run(context.Background(), "file")
}
```
