package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/p2pquake/stations_watcher/internal/retreiver"
)

func main() {
	out := flag.String("o", "Stations.csv", "output CSV path")
	flag.Parse()

	ctx := context.Background()
	stations, err := retreiver.RetreiveAndParse(ctx)
	if err != nil {
		fmt.Fprintln(os.Stderr, "fetch error:", err)
		os.Exit(1)
	}

	rows := make([]string, 0, len(stations))
	for i := range stations {
		s := &stations[i]
		rows = append(rows, fmt.Sprintf("%s,%s,%s,%s,%s",
			s.PrefName(), s.Name, s.Lat, s.Lon, s.AffiName()))
	}
	csv := strings.Join(rows, "\r\n")

	if err := os.WriteFile(*out, []byte(csv), 0o644); err != nil {
		fmt.Fprintln(os.Stderr, "write error:", err)
		os.Exit(1)
	}

	fmt.Printf("fetched %d stations -> %s\n", len(stations), *out)
}
