package retreiver

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"strconv"
)

const stationsURL = "https://www.jma.go.jp/jma/kishou/know/jishin/intens-st/stations.json"

type SeismicIntensityStation struct {
	Lat  StringNum `json:"lat"`
	Lon  StringNum `json:"lon"`
	Name string    `json:"name"`
	Pref string    `json:"pref"`
	Affi string    `json:"affi"`
}

// StringNum accepts either a JSON string or number, rounds to 2 decimal
// places, and stores the result as a fixed 2-decimal string (with trailing
// zeros). Rounding fixes float-precision artifacts in the upstream feed
// (e.g. 131.23499999...).
type StringNum string

func (s *StringNum) UnmarshalJSON(data []byte) error {
	var f float64
	if len(data) > 0 && data[0] == '"' {
		var v string
		if err := json.Unmarshal(data, &v); err != nil {
			return err
		}
		parsed, err := strconv.ParseFloat(v, 64)
		if err != nil {
			// Non-numeric string: keep as-is.
			*s = StringNum(v)
			return nil
		}
		f = parsed
	} else if err := json.Unmarshal(data, &f); err != nil {
		return err
	}
	*s = StringNum(strconv.FormatFloat(round2(f), 'f', 2, 64))
	return nil
}

func round2(f float64) float64 {
	return math.Round(f*100) / 100
}

func (s StringNum) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(s))
}

func (s *SeismicIntensityStation) Line() string {
	return fmt.Sprintf("%s,%s,%s,%s,%s", s.Lat, s.Lon, s.Name, s.Pref, s.Affi)
}

func (s *SeismicIntensityStation) PrefName() string {
	prefs := map[string]string{
		"1": "北海道", "2": "青森県", "3": "岩手県", "4": "宮城県", "5": "秋田県",
		"6": "山形県", "7": "福島県", "8": "茨城県", "9": "栃木県", "10": "群馬県",
		"11": "埼玉県", "12": "千葉県", "13": "東京都", "14": "神奈川県", "15": "新潟県",
		"16": "富山県", "17": "石川県", "18": "福井県", "19": "山梨県", "20": "長野県",
		"21": "岐阜県", "22": "静岡県", "23": "愛知県", "24": "三重県", "25": "滋賀県",
		"26": "京都府", "27": "大阪府", "28": "兵庫県", "29": "奈良県", "30": "和歌山県",
		"31": "鳥取県", "32": "島根県", "33": "岡山県", "34": "広島県", "35": "山口県",
		"36": "徳島県", "37": "香川県", "38": "愛媛県", "39": "高知県", "40": "福岡県",
		"41": "佐賀県", "42": "長崎県", "43": "熊本県", "44": "大分県", "45": "宮崎県",
		"46": "鹿児島県", "47": "沖縄県",
	}
	if name, ok := prefs[s.Pref]; ok {
		return name
	}
	return "不明"
}

func (s *SeismicIntensityStation) AffiName() string {
	switch s.Affi {
	case "0":
		return "気象庁"
	case "1":
		return "地方公共団体"
	default:
		return "防災科研"
	}
}

func RetreiveAndParse(ctx context.Context) ([]SeismicIntensityStation, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, stationsURL, nil)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var items []SeismicIntensityStation
	if err := json.Unmarshal(body, &items); err != nil {
		return nil, err
	}
	return items, nil
}
