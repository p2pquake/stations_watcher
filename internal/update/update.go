package update

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/p2pquake/stations_watcher/internal/github"
	"github.com/p2pquake/stations_watcher/internal/retreiver"
	"github.com/p2pquake/stations_watcher/internal/storage"
)

const bucketName = "p2pquake-seismic-intensity-stations"

// Update fetches the latest stations, compares against stored copy, and on
// difference saves new data + creates a GitHub issue. Returns true if there
// was a change (matches Rust implementation's exit-code semantics).
func Update(ctx context.Context) (bool, error) {
	store, err := storage.NewS3Storage(ctx, bucketName)
	if err != nil {
		return false, fmt.Errorf("init storage: %w", err)
	}

	oldStations, err := store.Load(ctx)
	if err != nil {
		return false, fmt.Errorf("load: %w", err)
	}
	stations, err := retreiver.RetreiveAndParse(ctx)
	if err != nil {
		return false, fmt.Errorf("retreive: %w", err)
	}

	oldStr := joinLines(oldStations)
	newStr := joinLines(stations)

	lds := lineDiffs(oldStr, newStr)
	if !hasChange(lds) {
		fmt.Println("no diff.")
		return false, nil
	}

	if err := store.Save(ctx, stations); err != nil {
		return false, fmt.Errorf("save json: %w", err)
	}

	diffText := unifiedDiff(lds, 1)
	fmt.Print(diffText)

	csv := buildCSV(stations)
	if err := store.SaveCSV(ctx, csv); err != nil {
		return false, fmt.Errorf("save csv: %w", err)
	}

	if err := github.CreateIssue(ctx, github.CreateIssueParams{
		Org:            os.Getenv("org"),
		Repo:           os.Getenv("repo"),
		AppID:          os.Getenv("app_id"),
		PemBucket:      os.Getenv("pem_bucket"),
		PemKey:         os.Getenv("pem_key"),
		InstallationID: os.Getenv("installation_id"),
		Diff:           diffText,
	}); err != nil {
		return true, fmt.Errorf("create issue: %w", err)
	}

	return true, nil
}

func joinLines(stations []retreiver.SeismicIntensityStation) string {
	var b strings.Builder
	for i := range stations {
		b.WriteString(stations[i].Line())
		b.WriteString("\n")
	}
	return b.String()
}

func buildCSV(stations []retreiver.SeismicIntensityStation) string {
	rows := make([]string, 0, len(stations))
	for i := range stations {
		s := &stations[i]
		rows = append(rows, fmt.Sprintf("%s,%s,%s,%s,%s",
			s.PrefName(), s.Name, s.Lat, s.Lon, s.AffiName()))
	}
	return strings.Join(rows, "\r\n")
}
