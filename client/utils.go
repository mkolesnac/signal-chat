package main

import "fmt"

func panicIfEmpty(argName, value string) {
	if len(value) == 0 {
		panic(fmt.Sprintf("%s cannot be empty", argName))
	}
}
