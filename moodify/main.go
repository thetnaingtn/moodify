package main

import (
	"context"

	"github.com/thetnaingtn/moodify/cmd"
)

var (
	version = "v0.1.0"
)

func main() {
	ctx := context.Background()
	cmd.Execute(ctx)
}
