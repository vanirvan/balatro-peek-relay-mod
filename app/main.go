package main

import (
	"app/analyzer"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/egoist/mygo"
)

type Card struct {
	Rank        string      `json:"rank"`
	Suit        string      `json:"suit"`
	ID          interface{} `json:"id"`
	Seal        *string     `json:"seal,omitempty"`
	Enhancement *string     `json:"enhancement,omitempty"`
	Edition     *string     `json:"edition,omitempty"`
	Debuffed    bool        `json:"debuffed"`
	Selected    bool        `json:"selected"`
	Scoring     bool        `json:"scoring"`
}

type SpecialCard struct {
	ID         interface{} `json:"id"`
	Name       string      `json:"name"`
	Set        string      `json:"set"`
	PosX       int         `json:"pos_x"`
	PosY       int         `json:"pos_y"`
	Edition    *string     `json:"edition,omitempty"`
	Debuffed   bool        `json:"debuffed"`
	Rarity     int         `json:"rarity"`
	Eternal    bool        `json:"eternal"`
	Perishable bool        `json:"perishable"`
	Rental     bool        `json:"rental"`
	Pinned     bool        `json:"pinned"`
	ValMult    *float64    `json:"val_mult,omitempty"`
	ValXMult   *float64    `json:"val_x_mult,omitempty"`
	ValChips   *float64    `json:"val_chips,omitempty"`
}

type VoucherItem struct {
	Key   string `json:"key"`
	Name  string `json:"name"`
	Set   string `json:"set"`
	PosX  int    `json:"pos_x"`
	PosY  int    `json:"pos_y"`
	Order int    `json:"order"`
}

type ScoringHand struct {
	Name  string `json:"name"`
	Level int    `json:"level"`
	Chips int    `json:"chips"`
	Mult  int    `json:"mult"`
}

type LiveState struct {
	Timestamp    int64         `json:"timestamp"`
	Status       string        `json:"status"`
	Seed         string        `json:"seed"`
	Ante         int           `json:"ante"`
	Round        int           `json:"round"`
	Blind        string        `json:"blind"`
	BlindChips   int64         `json:"blind_chips"`
	HandsLeft    int           `json:"hands_left"`
	DiscardsLeft int           `json:"discards_left"`
	DeckSize     int           `json:"deck_size"`
	ScoringHand  ScoringHand   `json:"scoring_hand"`
	Jokers       []SpecialCard `json:"jokers"`
	Consumables  []SpecialCard `json:"consumables"`
	Vouchers     []VoucherItem `json:"vouchers"`
	DrawPile     []Card        `json:"draw_pile"`
	Hand         []Card        `json:"hand"`
}

type PeekBridge struct{}

func getLiveJSONPath() string {
	appData := os.Getenv("APPDATA")
	return filepath.Join(appData, "Balatro", "Mods", "PeekRelay", "live_deck.json")
}

func readLiveState() (*LiveState, string, error) {
	path := getLiveJSONPath()
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, "", err
	}
	h := sha256.Sum256(data)
	hash := hex.EncodeToString(h[:])

	var state LiveState
	if err := json.Unmarshal(data, &state); err != nil {
		log.Printf("json unmarshal error: %v", err)
		return nil, "", err
	}
	return &state, hash, nil
}

// GetCurrentState returns the latest state immediately when requested by frontend.
func (PeekBridge) GetCurrentState() *LiveState {
	state, _, _ := readLiveState()
	return state
}

// GetPredictions computes or returns predicted future antes/vouchers/bosses based on live state.
func (PeekBridge) GetPredictions(seed string, ante int, vouchers []string) *analyzer.PredictionResult {
	if strings.TrimSpace(seed) == "" {
		return nil
	}
	return analyzer.AnalyzeSeed(seed, ante, vouchers, 8)
}

// StateUpdated is emitted to frontend whenever live_deck.json changes.
var StateUpdated = mygo.NewEvent[LiveState]("state_updated")

// PredictionUpdated is emitted to frontend whenever seed, ante, or inventory changes.
var PredictionUpdated = mygo.NewEvent[analyzer.PredictionResult]("prediction_updated")

func main() {
	mygo.Bind(PeekBridge{})

	mygo.App.WhenReady(func() {
		win := mygo.NewWindow(mygo.WindowOptions{
			Title:           "Peek Deck",
			Width:           960,
			Height:          680,
			MinWidth:        520,
			MinHeight:       400,
			BackgroundColor: "#181a20",
			StateKey:        "peekdeck-main",
			URL:             "/",
		})

		go func() {
			var lastHash string
			var lastSeed string
			var lastAnte int
			var lastVoucherCount int
			var lastJokerCount int
			var lastConsumableCount int

			ticker := time.NewTicker(200 * time.Millisecond)
			defer ticker.Stop()

			for range ticker.C {
				state, hash, err := readLiveState()
				if err != nil || state == nil {
					continue
				}
				if hash != lastHash {
					lastHash = hash
					if StateUpdated.Emit(win, *state) != nil {
						return // window closed
					}

					// Retrigger seed analyzer efficiently whenever:
					// 1. Seed changes
					// 2. Ante changes
					// 3. Voucher is bought/redeemed
					// 4. Joker or Consumable count changes
					curSeed := strings.TrimSpace(state.Seed)
					vCount := len(state.Vouchers)
					jCount := len(state.Jokers)
					cCount := len(state.Consumables)

					if curSeed != "" && (curSeed != lastSeed || state.Ante != lastAnte || vCount != lastVoucherCount || jCount != lastJokerCount || cCount != lastConsumableCount) {
						lastSeed = curSeed
						lastAnte = state.Ante
						lastVoucherCount = vCount
						lastJokerCount = jCount
						lastConsumableCount = cCount

						vKeys := make([]string, 0, vCount)
						for _, v := range state.Vouchers {
							vKeys = append(vKeys, v.Key)
						}

						pred := analyzer.AnalyzeSeed(curSeed, state.Ante, vKeys, 8)
						if pred != nil {
							_ = PredictionUpdated.Emit(win, *pred)
						}
					}
				}
			}
		}()
	})

	if err := mygo.App.Run(); err != nil {
		log.Fatal(err)
	}
}
