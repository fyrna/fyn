package main

import (
	"context"

	"github.com/fyrna/x/sh"

	"github.com/fyrna/fn/cli"
	"github.com/fyrna/fn/task"
)

func main() {
	t := task.New()

	t.Unit("echo", func(ctx context.Context) error {
		return sh.S(`echo "im cutee"`)
	})

	t.Unit("setup-test", func(ctx context.Context) error {
		return sh.S(`echo "//go:build ignore
	                package main
                        import \"fmt\"

                        func main() {
                                fmt.Println(\"miaw >w<\")
                        }" > myprogram.go`)
	})

	t.Unit("fish", func(ctx context.Context) error {
		sh.Exec(ctx, &sh.Options{
			Shell: "fish",
		}, `
echo "Nyaa~ mulai testing :3"

# Cek apakah Go terinstal
if type -q go
    echo "Go terdeteksi, versi:" (go version)
else
    echo "Go tidak ditemukan di Termux (〒﹏〒)"
end

# Cek apakah file myprogram ada dan bisa dijalankan
if test -x ./myprogram
    echo "Menjalankan ./myprogram nya~ >w<"
    set output (./myprogram ^/dev/null)
    if test $status -eq 0
        echo "Program berhasil dijalankan!"
        echo "Output: $output"
    else
        echo "Program gagal dijalankan (╥﹏╥)"
    end
else if test -f ./myprogram.go
    echo "Ditemukan myprogram.go, coba build dulu yaa~"
    go build -o myprogram myprogram.go
    if test -x ./myprogram
        echo "Build sukses! Menjalankan program~"
        ./myprogram
    else
        echo "Build gagal... (>_<)"
    end
else
    echo "Tidak ada ./myprogram atau ./myprogram.go di sini nyaaa~"
end

# Cek apakah config.yaml ada
if test -f config.yaml
    echo "config.yaml ditemukan!"
else
    echo "config.yaml tidak ada..."
end

echo "Testing selesai~ nyaaa~ ฅ^•ﻌ•^ฅ"`)

		return nil
	})

	t.Unit("cleanup", func(ctx context.Context) error {
		return sh.S("rm -rf myprogram myprogram.go")
	})

	t.Unit("test:fish", t.Series("setup-test", "fish", "cleanup"))
	t.Unit("test:all", t.Series("echo", "test:fish"))

	cli.Run(t)
}
