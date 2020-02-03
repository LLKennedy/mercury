package main

import (
	fmt "fmt"
	"net/http"
	"net/url"

	"github.com/LLKennedy/httpgrpc/internal/testservice/proxy"
	"github.com/LLKennedy/httpgrpc/internal/testservice/service"
)

func main() {
	s := service.New()
	startErr1 := make(chan error)
	startErr2 := make(chan error)
	go startServer(s, startErr1)

	client, err := s.MakeClientConn()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	p := proxy.New(client)
	go startServer(p, startErr2)

	// time.Sleep(1 * time.Second)
	fmt.Println("Sending HTTP request")

	webClient := &http.Client{}
	req := &http.Request{
		Method: http.MethodGet,
		URL: &url.URL{
			Host:   "localhost:4848",
			Scheme: "http",
		},
	}
	res, err := webClient.Do(req)
	if err != nil {
		fmt.Printf("error sending web request: %v\n", err)
	} else {
		fmt.Printf("http response: %v\n", res)
	}

	select {
	case err = <-startErr1:
	case err = <-startErr2:
	}
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
