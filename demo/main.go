package main

import (
	_ "github.com/cpunion/clibs/wasi-libc/v25"
	"github.com/goplus/llgo/c"
)

func main() {
	c.Printf(c.Str("Hello %s\n"), c.Str("world"))
}
