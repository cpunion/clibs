module github.com/cpunion/clibs/demo

go 1.20

require (
	github.com/cpunion/clibs/wasi-libc/v25 v25.0.6
	github.com/goplus/llgo v0.10.2-0.20250324105426-c968d8ca2e99
)

replace github.com/cpunion/clibs/wasi-libc/v25 => ../wasi-libc
