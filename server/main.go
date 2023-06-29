package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"sync"
)

func main() {
	if err := run(); err != nil {
		fmt.Printf("error: %v\n", err)
	}
}

func run() error {
	if len(os.Args) < 2 {
		return errors.New("please provide a path to client binary")
	}

	fmt.Printf("pid: %d\n", os.Getpid())
	cmds := spawn(os.Args[1], 50)

	for _, c := range cmds {
		if err := c.Wait(); err != nil {
			return fmt.Errorf("cmd wait: %w", err)
		}
	}
	return nil
}

func spawn(binaryPath string, n int) []*exec.Cmd {
	cmds := make([]*exec.Cmd, n)
	for i := 0; i < n; i++ {
		cmds[i] = exec.Command(binaryPath)
	}

	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		cmd := cmds[i]
		go func() {
			if err := start(cmd); err != nil {
				panic("TODO")
			}
			wg.Done()
		}()
	}
	wg.Wait()
	return cmds
}

func start(cmd *exec.Cmd) error {
	_, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("create stdout pipe: %w", err)
	}
	_, err = cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("create stdin pipe: %w", err)
	}
	if err = cmd.Start(); err != nil {
		return fmt.Errorf("run cmd: %w", err)
	}
	return nil
}
