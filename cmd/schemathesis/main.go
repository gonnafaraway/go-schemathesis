package main

import (
	"log"
	"os"

	"github.com/gonnafaraway/go-schemathesis/internal/cli"
)

func main() {
	if err := cli.Execute(os.Args[1:]); err != nil {
		log.Fatal(err)
	}
}
