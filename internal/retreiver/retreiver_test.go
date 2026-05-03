package retreiver

import (
	"encoding/json"
	"testing"
)

func TestStringNumRounding(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{"float artifact rounds up", `131.23499999999999`, "131.235"},
		{"already short float", `43.17`, "43.17"},
		{"string number", `"141.32"`, "141.32"},
		{"string with float artifact", `"131.23499999999999"`, "131.235"},
		{"integer float", `35`, "35"},
		{"negative float", `-0.0004`, "-0"},
		{"rounds half away from zero", `1.2345`, "1.235"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var s StringNum
			if err := json.Unmarshal([]byte(tc.in), &s); err != nil {
				t.Fatalf("unmarshal: %v", err)
			}
			if string(s) != tc.want {
				t.Fatalf("got %q want %q", s, tc.want)
			}
		})
	}
}

func TestStationUnmarshal(t *testing.T) {
	raw := `[{"lat":"43.17","lon":131.23499999999999,"name":"x","pref":"1","affi":"0"}]`
	var stations []SeismicIntensityStation
	if err := json.Unmarshal([]byte(raw), &stations); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if string(stations[0].Lat) != "43.17" {
		t.Fatalf("lat: %q", stations[0].Lat)
	}
	if string(stations[0].Lon) != "131.235" {
		t.Fatalf("lon: %q", stations[0].Lon)
	}
}
