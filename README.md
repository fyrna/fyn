# fyrna/fytask

i want to build my own task runner, but i realize it fking hard and **WORTHLESS** to designing new domain spesific language JUST FOR **STUPID PROJECT**

so, i think, can i create my own inside Go?

### quick example

```go
//go:build ignore
package main

import (
	"context"

	fy "github.com/fyrna/fytask"
)

func main() {
	t := fy.New()

	t.Task("file", func(ctx context.Context) error {
		fy.Sh(ctx, `echo "miaw" > file.txt`)
		fy.Sh(ctx, "cat file.txt")
		fy.Sh(ctx, "rm -f file.txt")

		fy.Execute(ctx,
			"for i in (seq 1 3); echo $i; end",
			&fy.ShellOptions{
				Shell: "fish",
			})

		return nil
  })

	t.Run(context.Background(), "file")
}
```
