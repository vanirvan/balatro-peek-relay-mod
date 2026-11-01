package analyzer

// AtlasPos stores sprite atlas grid coordinates.
type AtlasPos struct {
	X int `json:"x"`
	Y int `json:"y"`
}

// CardStickers tracks stake modifiers.
type CardStickers struct {
	Eternal    bool `json:"eternal"`
	Perishable bool `json:"perishable"`
	Rental     bool `json:"rental"`
}

// CardItem represents any card rendered in packs or tag rewards.
type CardItem struct {
	Name        string       `json:"name"`
	Set         string       `json:"set"` // "Joker", "Tarot", "Planet", "Spectral", "Playing Card"
	PosX        int          `json:"pos_x"`
	PosY        int          `json:"pos_y"`
	Edition     string       `json:"edition,omitempty"`
	Stickers    CardStickers `json:"stickers"`
	Rank        string       `json:"rank,omitempty"`
	Suit        string       `json:"suit,omitempty"`
	Enhancement string       `json:"enhancement,omitempty"`
	Seal        string       `json:"seal,omitempty"`
}

// PackPrediction represents one booster pack and the cards inside it.
type PackPrediction struct {
	Name    string     `json:"name"`    // e.g. "Jumbo Buffoon Pack"
	Type    string     `json:"type"`    // e.g. "Buffoon Pack"
	Size    int        `json:"size"`    // e.g. 4
	Choices int        `json:"choices"` // e.g. 1
	Title   string     `json:"title"`   // e.g. "Buffoon 4 cards, pick 1"
	Cards   []CardItem `json:"cards"`
}

// BlindPacks groups packs available after each blind shop.
type BlindPacks struct {
	Blind string           `json:"blind"` // "Small Blind", "Big Blind", "Boss Blind"
	Packs []PackPrediction `json:"packs"`
}

// BossRef is a lightweight reference to a boss blind.
type BossRef struct {
	Name string `json:"name"`
	Key  string `json:"key"`
	PosX int    `json:"pos_x"`
	PosY int    `json:"pos_y"`
}

// BossPrediction contains the default boss and possible rerolls.
type BossPrediction struct {
	Name               string    `json:"name"`
	Key                string    `json:"key"`
	PosX               int       `json:"pos_x"`
	PosY               int       `json:"pos_y"`
	DirectorCutReroll  *BossRef  `json:"director_cut_reroll,omitempty"`
	RetconRerolls      []BossRef `json:"retcon_rerolls,omitempty"`
}

// VoucherPrediction holds voucher info.
type VoucherPrediction struct {
	Key      string `json:"key"`
	Name     string `json:"name"`
	PosX     int    `json:"pos_x"`
	PosY     int    `json:"pos_y"`
	Resample int    `json:"resample"`
}

// TagPrediction represents a skip tag and its predicted reward.
type TagPrediction struct {
	Name          string             `json:"name"`
	Blind         string             `json:"blind"` // "Small Blind" or "Big Blind"
	PosX          int                `json:"pos_x"`
	PosY          int                `json:"pos_y"`
	RewardType    string             `json:"reward_type"` // "voucher", "pack", "joker", "hand", "bonus", "none"
	RewardTitle   string             `json:"reward_title"`
	RewardCards   []CardItem         `json:"reward_cards,omitempty"`
	RewardVoucher *VoucherPrediction `json:"reward_voucher,omitempty"`
	RewardHand    string             `json:"reward_hand,omitempty"`
}

// AntePrediction groups all predictions for a single ante.
type AntePrediction struct {
	Ante    int               `json:"ante"`
	Boss    BossPrediction    `json:"boss"`
	Voucher VoucherPrediction `json:"voucher"`
	Tags    []TagPrediction   `json:"tags"`
	Packs   []BlindPacks      `json:"packs"`
}

// PredictionResult is the complete payload sent to the frontend.
type PredictionResult struct {
	CurrentAnte int              `json:"current_ante"`
	Seed        string           `json:"seed"`
	Deck        string           `json:"deck"`
	Stake       int              `json:"stake"`
	Antes       []AntePrediction `json:"antes"`
}
