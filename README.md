
This small codebase reproduces an issue with `forkExec` in Go1.21rc2.
Specifically, when concurrently spinning up a bunch of child processes
that communicate via `os/exec.Command` with stdin/stdout pipes, the program 
will often hang indefinitely in `forkExec`.

This only seems to happen on Go 1.21 on apple silicon.

1. Build the "toy plugin" binary

```bash
$ cd plugin
$ go build .
```

2. Run the "server", pointing to the toy plugin binary (multiple times to reproduce)
```bash
$ cd ../server
$ go build .
$ ./server ../plugin/toy_plugin
pid: 40584
```

This will sometimes cause the program to hang indefinitely.

When attaching to the process via delve, the main goroutine is waiting
on the wait group as expected, but one of the goroutines spun up in `spawn`
is hanging in `forkExec`:

```bash
$ dlv attach 40584
Type 'help' for list of commands.
(dlv) grs
  Goroutine 1 - User: /Users/joaks/go/src/github.com/golang/go/src/runtime/sema.go:62 sync.runtime_Semacquire (0x1026bf57c) [semacquire]
  Goroutine 2 - User: /Users/joaks/go/src/github.com/golang/go/src/runtime/proc.go:399 runtime.gopark (0x102695198) [force gc (idle)]
  Goroutine 3 - User: /Users/joaks/go/src/github.com/golang/go/src/runtime/proc.go:399 runtime.gopark (0x102695198) [GC sweep wait]
  Goroutine 4 - User: /Users/joaks/go/src/github.com/golang/go/src/runtime/proc.go:399 runtime.gopark (0x102695198) [GC scavenge wait]
  Goroutine 5 - User: /Users/joaks/go/src/github.com/golang/go/src/runtime/proc.go:399 runtime.gopark (0x102695198) [finalizer wait]
  Goroutine 12 - User: /Users/joaks/go/src/github.com/golang/go/src/runtime/sys_darwin.go:24 syscall.syscall (0x1026bfaf8) (thread 18311682) [timer goroutine (idle)]
[6 goroutines]
(dlv) gr 12
Switched from 0 to 12 (thread 18311682)
(dlv) stack
 0  0x000000018f884acc in ???
    at ?:-1
 1  0x00000001026c0b58 in runtime.systemstack_switch
    at /Users/joaks/go/src/github.com/golang/go/src/runtime/asm_arm64.s:200
 2  0x00000001026b19dc in runtime.libcCall
    at /Users/joaks/go/src/github.com/golang/go/src/runtime/sys_libc.go:49
 3  0x00000001026bfaf8 in syscall.syscall
    at /Users/joaks/go/src/github.com/golang/go/src/runtime/sys_darwin.go:24
 4  0x00000001026daa5c in syscall.readlen
    at /Users/joaks/go/src/github.com/golang/go/src/syscall/syscall_darwin.go:242
 5  0x00000001026d9c30 in syscall.forkExec
    at /Users/joaks/go/src/github.com/golang/go/src/syscall/exec_unix.go:217
 6  0x00000001026e9628 in syscall.StartProcess
    at /Users/joaks/go/src/github.com/golang/go/src/syscall/exec_unix.go:334
 7  0x00000001026e9628 in os.startProcess
    at /Users/joaks/go/src/github.com/golang/go/src/os/exec_posix.go:54
 8  0x00000001026e9340 in os.StartProcess
    at /Users/joaks/go/src/github.com/golang/go/src/os/exec.go:111
 9  0x00000001026fc534 in os/exec.(*Cmd).Start
    at /Users/joaks/go/src/github.com/golang/go/src/os/exec/exec.go:693
10  0x00000001026ff368 in main.(*client).start
    at ./server/main.go:105
11  0x00000001026fefa8 in main.spawn.func1
    at ./server/main.go:46
12  0x00000001026c3024 in runtime.goexit
    at /Users/joaks/go/src/github.com/golang/go/src/runtime/asm_arm64.s:1197
```

`git bisect` for this issue led me to this commit:
https://go-review.googlesource.com/c/go/+/421441
