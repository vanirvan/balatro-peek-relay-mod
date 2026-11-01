package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"app/analyzer"

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

type PoolFlagsMap map[string]bool

func (p *PoolFlagsMap) UnmarshalJSON(data []byte) error {
	trimmed := strings.TrimSpace(string(data))
	if trimmed == "[]" || trimmed == "null" || trimmed == "" {
		*p = make(map[string]bool)
		return nil
	}
	var m map[string]bool
	if err := json.Unmarshal(data, &m); err != nil {
		*p = make(map[string]bool)
		return nil
	}
	*p = m
	return nil
}

type LiveState struct {
	Timestamp    int64         `json:"timestamp"`
	Status       string        `json:"status"`
	Seed         string        `json:"seed"`
	Ante         int           `json:"ante"`
	Round        int           `json:"round"`
	Deck         string        `json:"deck"`
	Stake        int           `json:"stake"`
	PoolFlags    PoolFlagsMap  `json:"pool_flags"`
	Blind        string        `json:"blind"`
	BlindChips   int64           `json:"blind_chips"`
	HandsLeft    int             `json:"hands_left"`
	DiscardsLeft int             `json:"discards_left"`
	DeckSize     int             `json:"deck_size"`
	ScoringHand  ScoringHand     `json:"scoring_hand"`
	Jokers       []SpecialCard   `json:"jokers"`
	Consumables  []SpecialCard   `json:"consumables"`
	Vouchers     []VoucherItem   `json:"vouchers"`
	DrawPile     []Card          `json:"draw_pile"`
	Hand         []Card          `json:"hand"`
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

// GetPredictions executes seed prediction on demand from the frontend.
func (PeekBridge) GetPredictions(
	seed string,
	ante int,
	vouchers []string,
	jokers []string,
	consumables []string,
	poolFlags map[string]bool,
	deck string,
	stake int,
) *analyzer.PredictionResult {
	if strings.TrimSpace(seed) == "" {
		return nil
	}
	return analyzer.AnalyzeSeed(seed, ante, vouchers, jokers, consumables, poolFlags, deck, stake, 16)
}

// StateUpdated is emitted to frontend whenever live_deck.json changes.
var StateUpdated = mygo.NewEvent[LiveState]("state_updated")

// PredictionUpdated is emitted to frontend whenever prediction inputs change.
var PredictionUpdated = mygo.NewEvent[analyzer.PredictionResult]("prediction_updated")

func computePredictionSignature(state *LiveState) string {
	if state == nil || strings.TrimSpace(state.Seed) == "" {
		return ""
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("s:%s|a:%d|d:%s|st:%d|", state.Seed, state.Ante, state.Deck, state.Stake))

	// Vouchers
	sb.WriteString("v:")
	for _, v := range state.Vouchers {
		sb.WriteString(v.Key)
		sb.WriteByte(',')
	}
	sb.WriteByte('|')

	// Jokers
	sb.WriteString("j:")
	for _, j := range state.Jokers {
		sb.WriteString(j.Name)
		sb.WriteByte(',')
	}
	sb.WriteByte('|')

	// Consumables
	sb.WriteString("c:")
	for _, c := range state.Consumables {
		sb.WriteString(c.Name)
		sb.WriteByte(',')
	}
	sb.WriteByte('|')

	// Unlock conditions (PoolFlags)
	sb.WriteString("pf:")
	keys := []string{
		"glass_joker_unlocked",
		"gros_michel_extinct",
		"cavendish_unlocked",
		"stone_joker_unlocked",
		"steel_joker_unlocked",
		"lucky_cat_unlocked",
		"golden_ticket_unlocked",
		"planet_x_unlocked",
		"ceres_unlocked",
		"eris_unlocked",
	}
	for _, k := range keys {
		if state.PoolFlags != nil && state.PoolFlags[k] {
			sb.WriteString(k)
			sb.WriteString("=1,")
		}
	}

	return sb.String()
}

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
			var lastPredSig string
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

					// Reactive check: only recalculate predictions if seed, ante, vouchers, jokers, consumables, or unlock flags changed
					predSig := computePredictionSignature(state)
					if predSig != "" && predSig != lastPredSig {
						lastPredSig = predSig

						vKeys := make([]string, 0, len(state.Vouchers))
						for _, v := range state.Vouchers {
							vKeys = append(vKeys, v.Key)
						}
						jNames := make([]string, 0, len(state.Jokers))
						for _, j := range state.Jokers {
							jNames = append(jNames, j.Name)
						}
						cNames := make([]string, 0, len(state.Consumables))
						for _, c := range state.Consumables {
							cNames = append(cNames, c.Name)
						}

						pred := analyzer.AnalyzeSeed(
							state.Seed,
							state.Ante,
							vKeys,
							jNames,
							cNames,
							map[string]bool(state.PoolFlags),
							state.Deck,
							state.Stake,
							16,
						)
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
