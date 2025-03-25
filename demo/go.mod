module github.com/goplus/clibs/demo

go 1.22

require (
	github.com/goplus/clibs/bdwgc v0.0.0-00010101000000-000000000000
	github.com/goplus/clibs/wasi-libc v0.0.0-00010101000000-000000000000
)

require github.com/goplus/llgo v0.10.1 // indirect

replace github.com/goplus/clibs => ../

replace github.com/goplus/clibs/bdwgc => ../bdwgc

replace github.com/goplus/clibs/wasi-libc => ../wasi-libc
