package main

import "fmt"

func requireNonEmpty(name, value string) {
	if len(value) == 0 {
		panic(fmt.Sprintf("%s cannot be empty", name))
	}
}
