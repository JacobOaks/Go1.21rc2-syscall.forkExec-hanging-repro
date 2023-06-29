package main

import (
	"bufio"
	"fmt"
	"os"
)

func main() {
	if err := relay(); err != nil {
		fmt.Printf("error: %v\n", err)
	}
}

func relay() error {
	reader := bufio.NewReader(os.Stdin)
	resp := make([]byte, 10)
	_, err := reader.Read(resp)
	if err != nil {
		return fmt.Errorf("read resp: %w", err)
	}
	fmt.Printf("plugin ack %v", string(resp))
	return nil
}
