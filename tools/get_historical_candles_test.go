package tools

import (
	"github.com/dvgamerr-app/go-bitkub/market"
	"reflect"
	"testing"
)

func TestLatestHistory(t *testing.T) {
	for _, tc := range []struct {
		name  string
		times []int64
		limit int
		want  []int64
	}{
		{"extra candles", []int64{1, 2, 3, 4}, 2, []int64{3, 4}},
		{"unordered candles", []int64{4, 1, 3, 2}, 2, []int64{3, 4}},
		{"short history", []int64{2, 1}, 3, []int64{1, 2}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			data := &market.HistoryResponse{Time: tc.times}
			for _, ts := range tc.times {
				v := float64(ts)
				data.Open = append(data.Open, v)
				data.High = append(data.High, v+1)
				data.Low = append(data.Low, v-1)
				data.Close = append(data.Close, v)
				data.Volume = append(data.Volume, v*10)
			}
			result, err := latestHistory(data, tc.limit)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(result.Time, tc.want) {
				t.Fatalf("timestamps=%v, want %v", result.Time, tc.want)
			}
			for i, ts := range result.Time {
				v := float64(ts)
				if result.Open[i] != v || result.High[i] != v+1 || result.Low[i] != v-1 || result.Close[i] != v || result.Volume[i] != v*10 {
					t.Fatal("OHLCV alignment lost")
				}
			}
		})
	}
	t.Run("reject mismatched arrays", func(t *testing.T) {
		if _, err := latestHistory(&market.HistoryResponse{Time: []int64{1}}, 2); err == nil {
			t.Fatal("expected error")
		}
	})
}
