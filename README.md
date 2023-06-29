
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

2. Run the "server", pointing to the toy plugin binary
```bash
$ cd ../server
$ go build .
$ ./server ../plugin/toy_plugin
pid: 40584
```

Very frequently this will cause the program to hang indefinitely.

When attaching to the process via delve, the main goroutine is waiting
on the wait group as expected, but one of the goroutines spun up in `spawn`
is hanging in `forkExec`:

```bash
$ dlv attach 40584
Type 'help' for list of commands.
(dlv) grs
  Goroutine 1 - User: ./go/src/github.com/golang/go/src/runtime/sema.go:62 sync.runtime_Semacquire (0x100926bec) [semacquire]
  Goroutine 2 - User: ./go/src/github.com/golang/go/src/runtime/proc.go:399 runtime.gopark (0x1008fcdc8) [force gc (idle)]
  Goroutine 3 - User: ./go/src/github.com/golang/go/src/runtime/proc.go:399 runtime.gopark (0x1008fcdc8) [GC sweep wait]
  Goroutine 4 - User: ./go/src/github.com/golang/go/src/runtime/proc.go:399 runtime.gopark (0x1008fcdc8) [GC scavenge wait]
  Goroutine 5 - User: ./go/src/github.com/golang/go/src/runtime/proc.go:399 runtime.gopark (0x1008fcdc8) [finalizer wait]
* Goroutine 15 - User: ./go/src/github.com/golang/go/src/runtime/sys_darwin.go:24 syscall.syscall (0x100927168) (thread 17924750) [timer goroutine (idle)]
[6 goroutines]
(dlv) gr 15
Switched from 15 to 15 (thread 17924750)
(dlv) stack
 0  0x000000018f884acc in ???
    at ?:-1
 1  0x00000001009281c8 in runtime.systemstack_switch
    at ./go/src/github.com/golang/go/src/runtime/asm_arm64.s:200
 2  0x000000010091914c in runtime.libcCall
    at ./go/src/github.com/golang/go/src/runtime/sys_libc.go:49
 3  0x0000000100927168 in syscall.syscall
    at ./go/src/github.com/golang/go/src/runtime/sys_darwin.go:24
 4  0x0000000100941fbc in syscall.readlen
    at ./go/src/github.com/golang/go/src/syscall/syscall_darwin.go:242
 5  0x0000000100941190 in syscall.forkExec
    at ./go/src/github.com/golang/go/src/syscall/exec_unix.go:217
 6  0x0000000100950bf8 in syscall.StartProcess
    at ./go/src/github.com/golang/go/src/syscall/exec_unix.go:334
 7  0x0000000100950bf8 in os.startProcess
    at ./go/src/github.com/golang/go/src/os/exec_posix.go:54
 8  0x0000000100950910 in os.StartProcess
    at ./go/src/github.com/golang/go/src/os/exec.go:109
 9  0x0000000100963a64 in os/exec.(*Cmd).Start
    at ./go/src/github.com/golang/go/src/os/exec/exec.go:693
10  0x00000001009665e0 in main.start
    at ./go/src/scratch/go121forkexec/server/main.go:63
11  0x00000001009664e8 in main.spawn.func1
    at ./go/src/scratch/go121forkexec/server/main.go:44
12  0x000000010092a694 in runtime.goexit
    at ./go/src/github.com/golang/go/src/runtime/asm_arm64.s:1197
```

`git bisect` for this issue led me to this commit:
https://go-review.googlesource.com/c/go/+/421441
