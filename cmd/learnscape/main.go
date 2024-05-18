package main

import (
	"context"
	"fmt"
	"os"

	i "github.com/dr0th3r/learnscape/internal"
	"github.com/dr0th3r/learnscape/internal/utils"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

const configPath = "./config/config.json"

func main() {
	ctx := context.Background()
	config, err := utils.ParseConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		return
	}
	if err := i.Run(ctx, config); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}
