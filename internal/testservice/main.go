package main

import (
	fmt "fmt"

	"github.com/LLKennedy/httpgrpc/internal/testservice/server"
)

func main() {
	s := server.New()
	startErr := make(chan error)
	go startServer(s, startErr)
	err := <-startErr
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Println("Closed server gracefully")
	}
}

type starter interface {
	Start() error
}

func startServer(s starter, out chan error) {
	defer func() {
		// just ignore panics
		recover()
	}()
	err := s.Start()
	if err != nil {
		out <- err
	}
}
