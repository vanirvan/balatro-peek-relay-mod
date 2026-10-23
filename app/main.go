package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"log"
	"os"
	"path/filepath"
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

type ScoringHand struct {
	Name  string `json:"name"`
	Level int    `json:"level"`
	Chips int    `json:"chips"`
	Mult  int    `json:"mult"`
}

type LiveState struct {
	Timestamp    int64       `json:"timestamp"`
	Status       string      `json:"status"`
	Seed         string      `json:"seed"`
	Ante         int         `json:"ante"`
	Round        int         `json:"round"`
	Blind        string      `json:"blind"`
	BlindChips   int64       `json:"blind_chips"`
	HandsLeft    int         `json:"hands_left"`
	DiscardsLeft int         `json:"discards_left"`
	DeckSize     int         `json:"deck_size"`
	ScoringHand  ScoringHand `json:"scoring_hand"`
	DrawPile     []Card      `json:"draw_pile"`
	Hand         []Card      `json:"hand"`
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

// StateUpdated is emitted to frontend whenever live_deck.json changes.
var StateUpdated = mygo.NewEvent[LiveState]("state_updated")

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
				}
			}
		}()
	})

	if err := mygo.App.Run(); err != nil {
		log.Fatal(err)
	}
}
