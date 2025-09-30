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

	t.Task("print", nil)
	t.Task("miaw", nil)
	t.Task("build", func(ctx context.Context) error {
		fy.Shell(ctx, "echo", "miaw :3")
		fy.Sh(ctx, `echo i try new abstraction`)

		return nil
	})

	t.Run(context.Background(), "build")
	t.List()
}
```
