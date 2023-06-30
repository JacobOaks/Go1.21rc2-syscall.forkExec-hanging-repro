package main

import (
	"errors"
	"fmt"
	"io"
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
	clients := spawn(os.Args[1], 50)

	// TODO: Use updated data...
	if _, err := sendAndReceive("some data", clients); err != nil {
		return err
	}
	stop(clients)

	return nil
}

func spawn(binaryPath string, n int) []*client {
	clients := make([]*client, n)
	for i := 0; i < n; i++ {
		clients[i] = newClient(binaryPath)
	}

	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		client := clients[i]
		go func() {
			if err := client.start(); err != nil {
				panic("TODO")
			}
			wg.Done()
		}()
	}
	wg.Wait()
	return clients
}

func sendAndReceive(data string, clients []*client) ([]string, error) {
	// We would parallelize this too, but keeping it simple for brevity.
	responses := make([]string, len(clients))
	for i, c := range clients {
		if _, err := c.stdin.Write([]byte(data)); err != nil {
			return nil, err
		}
		resp := make([]byte, 64)
		if _, err := c.stdout.Read(resp); err != nil {
			return nil, err
		}
		responses[i] = string(resp)
	}
	return responses, nil
}

func stop(clients []*client) error {
	// We would parallelize this too, but keeping it simple for brevity.
	for _, c := range clients {
		if err := c.stop(); err != nil {
			return err
		}
	}
	return nil
}

type client struct {
	cmd *exec.Cmd

	stdout io.ReadCloser
	stdin  io.WriteCloser
}

func newClient(binary string) *client {
	return &client{
		cmd: exec.Command(binary),
	}
}

func (c *client) start() error {
	var err error
	c.stdout, err = c.cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("create stdout pipe: %w", err)
	}
	c.stdin, err = c.cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("create stdin pipe: %w", err)
	}
	if err = c.cmd.Start(); err != nil {
		return fmt.Errorf("run cmd: %w", err)
	}
	return nil
}

func (c *client) stop() error {
	return c.cmd.Wait()
}
