package main

import (
	"context"

	"github.com/dnitsch/aws-cli-auth/cmd"
)

func main() {
	cmd.Execute(context.Background())
}
