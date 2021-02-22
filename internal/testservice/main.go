package main

import (
	fmt "fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"github.com/LLKennedy/httpgrpc/v2/internal/testservice/proxy"
	"github.com/LLKennedy/httpgrpc/v2/internal/testservice/service"
)

func main() {
	s, err := service.New()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	startErr1 := make(chan error)
	startErr2 := make(chan error)
	go startServer(s, startErr1)
	time.Sleep(time.Second)
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
			Path:   "/Random",
			Scheme: "http",
		},
	}
	res, err := webClient.Do(req)
	if err != nil {
		fmt.Printf("error sending web request: %v\n", err)
	} else {
		fmt.Printf("http response: %+v\n", res)
		body, _ := ioutil.ReadAll(res.Body)
		fmt.Printf("http response body: %s\n", body)
	}

	select {
	case err = <-startErr1:
	case err = <-startErr2:
	default:
		fmt.Println("Closing without server-start error")
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
