// Package main is the main package of the Falcula CLI
package main

import (
	"github.com/LucasAVasco/falcula"
)

func main() {
	err := falcula.StartCli()
	if err != nil {
		panic(err)
	}
}
