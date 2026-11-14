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
		{"duplicate timestamps", []int64{1, 2, 2, 3}, 3, []int64{1, 2, 3}},
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

func TestLatestHistoryKeepsLastUpdate(t *testing.T) {
	data := &market.HistoryResponse{Time: []int64{60, 60}, Open: []float64{100, 100}, High: []float64{101, 110}, Low: []float64{99, 95}, Close: []float64{101, 105}, Volume: []float64{1, 20}}
	got, err := latestHistory(data, 2)
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Time) != 1 || got.Close[0] != 105 || got.High[0] != 110 || got.Low[0] != 95 || got.Volume[0] != 20 {
		t.Fatalf("duplicate update not replaced: %+v", got)
	}
}

func TestClosedHistory(t *testing.T) {
	data := &market.HistoryResponse{Time: []int64{0, 60, 120, 180}, Open: []float64{1, 2, 3, 4}, High: []float64{1, 2, 3, 4}, Low: []float64{1, 2, 3, 4}, Close: []float64{1, 2, 3, 4}, Volume: []float64{10, 20, 30, 40}}
	for _, tc := range []struct {
		name    string
		now     int64
		want    []int64
		wantErr bool
	}{
		{"exclude forming candle", 190, []int64{60, 120}, false},
		{"include candle at close", 240, []int64{120, 180}, false},
		{"reject stale history", 400, nil, true},
		{"no completed candles", 30, nil, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := closedHistory(data, 2, 1, tc.now)
			if tc.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got.Time, tc.want) {
				t.Fatalf("times=%v want=%v", got.Time, tc.want)
			}
			if got.Close[len(got.Close)-1] != float64(got.Time[len(got.Time)-1]/60+1) {
				t.Fatal("candle fields misaligned")
			}
		})
	}
}

func TestPrepareClosedHistory(t *testing.T) {
	for _, tc := range []struct {
		name      string
		times     []int64
		now       int64
		want      []int64
		synthetic int
		wantErr   bool
	}{
		{"fill no trade intervals", []int64{0, 120}, 300, []int64{0, 60, 120, 180, 240}, 3, false},
		{"exclude forming candle", []int64{0, 60, 180}, 200, []int64{0, 60, 120}, 1, false},
		{"source outside lookback", []int64{0}, 1000, nil, 0, true},
		{"reject invalid grid", []int64{0, 90}, 200, nil, 0, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			data := &market.HistoryResponse{}
			for i, ts := range tc.times {
				v := 100 + float64(i)
				data.Time = append(data.Time, ts)
				data.Open = append(data.Open, v)
				data.High = append(data.High, v)
				data.Low = append(data.Low, v)
				data.Close = append(data.Close, v)
				data.Volume = append(data.Volume, 10)
			}
			got, actual, metadata, err := prepareClosedHistory(data, 5, 1, tc.now)
			if tc.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got.Time, tc.want) || metadata.SyntheticCount != tc.synthetic {
				t.Fatalf("timestamps=%v metadata=%+v", got.Time, metadata)
			}
			for i, ts := range got.Time {
				if !actual[ts] {
					if i == 0 || got.Volume[i] != 0 || got.Open[i] != got.Close[i-1] || got.High[i] != got.Close[i-1] || got.Low[i] != got.Close[i-1] || got.Close[i] != got.Close[i-1] {
						t.Fatal("invalid synthetic candle")
					}
				}
			}
			if len(data.Time) != len(tc.times) {
				t.Fatal("source modified")
			}
		})
	}
}
