package main

import (
	"fmt"

	"github.com/leandroluk/gox/di"
)

type Service interface {
	Do()
}

type ServiceImpl struct{}

func (s *ServiceImpl) Do() {
	fmt.Println("Service clean!")
}

func main() {
	fmt.Println("--- Enabling Debug Mode ---")
	di.Debug()

	fmt.Println("\n--- Registering Service ---")
	di.RegisterAs[Service](func() *ServiceImpl {
		return &ServiceImpl{}
	})

	fmt.Println("\n--- Resolving Service ---")
	svc := di.Resolve[Service]()
	svc.Do()

	fmt.Println("\n--- Triggering Error ---")
	// This should print the error before panicking
	di.Resolve[string]()
}
