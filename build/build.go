package build

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/fyrna/x/sh"
)

// Valid GOOS/GOARCH combinations supported by Go toolchain
const (
	// android
	TARGET_ANDROID_ARM64 = "android/arm64"
	TARGET_ANDROID_ARM   = "android/arm"

	// darwin
	TARGET_DARWIN_AMD64 = "darwin/amd64"
	TARGET_DARWIN_ARM64 = "darwin/arm64"
	TARGET_IOS_ARM64    = "ios/arm64"
	TARGET_IOS_AMD64    = "ios/amd64"

	// linux
	TARGET_LINUX_AMD64    = "linux/amd64"
	TARGET_LINUX_ARM64    = "linux/arm64"
	TARGET_LINUX_ARM      = "linux/arm"
	TARGET_LINUX_RISCV64  = "linux/riscv64"
	TARGET_LINUX_S390X    = "linux/s390x"
	TARGET_LINUX_PPC64LE  = "linux/ppc64le"
	TARGET_LINUX_MIPS64   = "linux/mips64"
	TARGET_LINUX_MIPS64LE = "linux/mips64le"
	TARGET_LINUX_LOONG64  = "linux/loong64"

	// windows
	TARGET_WINDOWS_AMD64 = "windows/amd64"
	TARGET_WINDOWS_ARM64 = "windows/arm64"
	TARGET_WINDOWS_386   = "windows/386"

	// freeBSD
	TARGET_FREEBSD_AMD64 = "freebsd/amd64"
	TARGET_FREEBSD_ARM64 = "freebsd/arm64"
	TARGET_FREEBSD_386   = "freebsd/386"

	// netBSD
	TARGET_NETBSD_AMD64 = "netbsd/amd64"
	TARGET_NETBSD_ARM64 = "netbsd/arm64"

	// openBSD
	TARGET_OPENBSD_AMD64 = "openbsd/amd64"
	TARGET_OPENBSD_ARM64 = "openbsd/arm64"

	// dragonflyBSD
	TARGET_DRAGONFLY_AMD64 = "dragonfly/amd64"

	// solaris & Illumos
	TARGET_SOLARIS_AMD64 = "solaris/amd64"
	TARGET_ILLUMOS_AMD64 = "illumos/amd64"

	// AIX
	TARGET_AIX_PPC64 = "aix/ppc64"

	// Plan9
	TARGET_PLAN9_AMD64 = "plan9/amd64"
	TARGET_PLAN9_386   = "plan9/386"

	// JS / WebAssembly
	TARGET_JS_WASM = "js/wasm"

	// WASI
	TARGET_WASI_WASM = "wasi/wasm"
)

// linux & android
// modern (common) architecture only !
func LinuxOnly() []string {
	return []string{
		TARGET_LINUX_AMD64,
		TARGET_LINUX_ARM64,
		TARGET_LINUX_ARM,
		TARGET_LINUX_S390X,
		TARGET_LINUX_RISCV64,

		TARGET_ANDROID_ARM64,
		TARGET_ANDROID_ARM,
	}
}

type Build struct {
	Targets []string
	Flags   []string
	Output  string
	Pattern string
	Mode    string
	EnvVars [][2]string
}

type Option interface {
	apply(*Build)
}

type optFn func(*Build)

func (f optFn) apply(b *Build) { f(b) }

func Setup(opt ...Option) *Build {
	b := &Build{}
	for _, a := range opt {
		a.apply(b)
	}
	return b
}

func Target(args ...any) Option {
	return optFn(func(b *Build) {
		for _, a := range args {
			switch v := a.(type) {
			case string:
				b.Targets = append(b.Targets, v)
			case []string:
				b.Targets = append(b.Targets, v...)
			default:
				// TODO: dont panic
				panic(fmt.Sprintf("invalid argument to build.Target(): %T", v))
			}
		}
	})
}

func GoFlag(flag string) Option {
	return optFn(func(b *Build) {
		b.Flags = append(b.Flags, flag)
	})
}

type TargetInfo struct {
	OS   string
	Arch string
}

func (b *Build) TargetInfo() []TargetInfo {
	var infos []TargetInfo
	for _, t := range b.Targets {
		os, arch, _ := strings.Cut(string(t), "/")
		infos = append(infos, TargetInfo{OS: os, Arch: arch})
	}
	return infos
}

func Output(name string) Option {
	return optFn(func(b *Build) {
		b.Output = name
	})
}

func OutputPattern(pattern string) Option {
	return optFn(func(b *Build) {
		b.Pattern = pattern
	})
}

// TODO: fix all *Flag* thing
func GoFlagKV(key, value string) Option {
	return optFn(func(b *Build) {
		b.Flags = append(b.Flags, key+"="+value)
	})
}

func Env(key, value string) Option {
	return optFn(func(b *Build) {
		b.EnvVars = append(b.EnvVars, [2]string{key, value})
	})
}

func Verbose() Option    { return GoFlag("-v") }
func Trimpath() Option   { return GoFlag("-trimpath") }
func Race() Option       { return GoFlag("-race") }
func RebuildAll() Option { return GoFlag("-a") }
func DryRun() Option     { return GoFlag("-n") }
func Work() Option       { return GoFlag("-work") }
func ModCacheRW() Option { return GoFlag("-modcacherw") }

func LdFlags(v string) Option   { return GoFlagKV("-ldflags", v) }
func GcFlags(v string) Option   { return GoFlagKV("-gcflags", v) }
func Mod(v string) Option       { return GoFlagKV("-mod", v) }
func P(n int) Option            { return GoFlagKV("-p", fmt.Sprint(n)) }
func BuildMode(v string) Option { return GoFlagKV("-buildmode", v) }

func Tags(tags ...string) Option {
	return GoFlagKV("-tags", strings.Join(tags, " "))
}

type Mode struct {
	Name  string
	Flags []Option
}

func DEBUG() Mode {
	return Mode{
		Name: "debug",
		Flags: []Option{
			Verbose(),
			GcFlags("all=-N -l"),
			Tags("debug"),
		},
	}
}

func RELEASE() Mode {
	return Mode{
		Name: "release",
		Flags: []Option{
			Trimpath(),
			LdFlags("-s -w"),
			Tags("release"),
		},
	}
}

func TINY() Mode {
	return Mode{
		Name: "tiny",
		Flags: []Option{
			Trimpath(),
			LdFlags("\"-s -w -buildid=\""), // yeah it needs manual ""
			Tags("tiny"),
		},
	}
}

func FAST() Mode {
	return Mode{
		Name: "fast",
		Flags: []Option{
			Trimpath(),
			Tags("fast"),
			GcFlags("all=-B"),
		},
	}
}

func OutputMode(m Mode) Option {
	return optFn(func(b *Build) {
		for _, f := range m.Flags {
			f.apply(b)
		}
		b.Mode = m.Name
	})
}

func Start(b *Build) error {
	for _, t := range b.TargetInfo() {
		out := b.Output
		if b.Pattern != "" {
			out = b.Pattern
			out = strings.ReplaceAll(out, "{os}", t.OS)
			out = strings.ReplaceAll(out, "{arch}", t.Arch)
			out = strings.ReplaceAll(out, "{mode}", b.Mode)
		}

		if out == "" {
			out = "app"
		}

		outPath := filepath.Clean(out)

		args := append([]string{"build"}, b.Flags...)
		args = append(args, "-o", outPath)

		fmt.Printf("🐾 Building for %s/%s → %s\n", t.OS, t.Arch, outPath)

		// Gabung jadi string
		cmd := fmt.Sprintf("go %s", strings.Join(args, " "))

		fmt.Printf("running %s\n\n", cmd)
		_, err := sh.Exec(context.Background(),
			&sh.Options{
				Env: b.buildEnv(t),
			},
			cmd,
		)

		if err != nil {
			return fmt.Errorf("build failed for %s/%s: %w", t.OS, t.Arch, err)
		}
	}
	return nil
}

func (b *Build) buildEnv(t TargetInfo) []string {
	env := []string{
		"GOOS=" + t.OS,
		"GOARCH=" + t.Arch,
	}
	for _, kv := range b.EnvVars {
		env = append(env, fmt.Sprintf("%s=%s", kv[0], kv[1]))
	}
	return env
}
