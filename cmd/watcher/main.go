package main

import (
	"context"
	"fmt"
	"os"

	"github.com/p2pquake/stations_watcher/internal/update"
)

func main() {
	changed, err := update.Update(context.Background())
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	if changed {
		os.Exit(1)
	}
	os.Exit(0)
}
