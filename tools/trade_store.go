package tools

import (
	"encoding/json"
	"os"
	"sync"
)

// ponytail: a personal trade journal holds dozens of rows, not millions — a flat
// JSON file with a global lock covers logging/history/expectancy without pulling in
// a SQLite engine + CGO. Upgrade path: swap loadTrades/saveTrades for a DB layer if
// the journal ever outgrows read-modify-write of one file.

type TradeRecord struct {
	ID         int     `json:"id"`
	Symbol     string  `json:"symbol"`
	Strategy   string  `json:"strategy,omitempty"`
	EntryDate  string  `json:"entry_date"`
	EntryPrice float64 `json:"entry_price"`
	Qty        float64 `json:"qty"`
	Stop       float64 `json:"stop,omitempty"`
	TP2R       float64 `json:"tp_2r,omitempty"`
	Reason     string  `json:"reason,omitempty"`
	Status     string  `json:"status"` // open | closed
	ExitDate   string  `json:"exit_date,omitempty"`
	ExitPrice  float64 `json:"exit_price,omitempty"`
	PnLTHB     float64 `json:"pnl_thb,omitempty"`
	PnLPct     float64 `json:"pnl_pct,omitempty"`
	RMultiple  float64 `json:"r_multiple,omitempty"`
	ExitReason string  `json:"exit_reason,omitempty"`
}

// ponytail: global lock, fine for single-user tool calls; shard per-symbol only if
// concurrency ever matters.
var storeMu sync.Mutex

func tradesFilePath() string {
	if p := os.Getenv("TRADES_FILE"); p != "" {
		return p
	}
	return "trades.json"
}

func loadTrades() ([]TradeRecord, error) {
	data, err := os.ReadFile(tradesFilePath())
	if err != nil {
		if os.IsNotExist(err) {
			return []TradeRecord{}, nil
		}
		return nil, err
	}
	if len(data) == 0 {
		return []TradeRecord{}, nil
	}
	var trades []TradeRecord
	if err := json.Unmarshal(data, &trades); err != nil {
		return nil, err
	}
	return trades, nil
}

func saveTrades(trades []TradeRecord) error {
	data, err := json.MarshalIndent(trades, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(tradesFilePath(), data, 0o644)
}

// addTrade appends a record (assigning the next id) and returns the new id.
func addTrade(rec TradeRecord) (int, error) {
	storeMu.Lock()
	defer storeMu.Unlock()

	trades, err := loadTrades()
	if err != nil {
		return 0, err
	}
	maxID := 0
	for _, t := range trades {
		if t.ID > maxID {
			maxID = t.ID
		}
	}
	rec.ID = maxID + 1
	trades = append(trades, rec)
	if err := saveTrades(trades); err != nil {
		return 0, err
	}
	return rec.ID, nil
}

// updateTrade applies mutate to the record with the given id and persists it.
func updateTrade(id int, mutate func(*TradeRecord)) (*TradeRecord, error) {
	storeMu.Lock()
	defer storeMu.Unlock()

	trades, err := loadTrades()
	if err != nil {
		return nil, err
	}
	for i := range trades {
		if trades[i].ID == id {
			mutate(&trades[i])
			if err := saveTrades(trades); err != nil {
				return nil, err
			}
			rec := trades[i]
			return &rec, nil
		}
	}
	return nil, nil // not found
}
