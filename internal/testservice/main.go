package main

import fmt "fmt"

func main() {
	s := NewServer()
	err := s.Start()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}
}
