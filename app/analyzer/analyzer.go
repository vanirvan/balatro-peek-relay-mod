package analyzer

import (
	"fmt"
)

// AntePrediction encapsulates predicted events for a specific ante.
type AntePrediction struct {
	Ante         int          `json:"ante"`
	Boss         string       `json:"boss"`
	BossKey      string       `json:"boss_key"`
	Voucher      VoucherPred  `json:"voucher"`
	SmallTag     string       `json:"small_tag"`
	BigTag       string       `json:"big_tag"`
	ShopVouchers []string     `json:"shop_vouchers"`
}

type VoucherPred struct {
	Key      string `json:"key"`
	Name     string `json:"name"`
	PosX     int    `json:"pos_x"`
	PosY     int    `json:"pos_y"`
	Resample int    `json:"resample"`
}

// PredictionResult is the payload sent to the desktop app.
type PredictionResult struct {
	CurrentAnte int              `json:"current_ante"`
	Seed        string           `json:"seed"`
	Antes       []AntePrediction `json:"antes"`
	NextTags    []string         `json:"next_voucher_tags"`
}

// AnalyzeSeed executes high-efficiency prediction using current live state.
// It runs in sub-millisecond time on CPU using zero allocs where possible.
func AnalyzeSeed(seed string, currentAnte int, redeemedVouchers []string, maxAnte int) *PredictionResult {
	if seed == "" {
		return nil
	}
	if maxAnte <= 0 {
		maxAnte = 8
	}
	if currentAnte > maxAnte {
		maxAnte = currentAnte + 2
	}

	rng := NewBalatroRNG(seed)

	usedVouchers := make(map[string]bool, 32)
	for _, v := range redeemedVouchers {
		usedVouchers[v] = true
	}

	bossesUsed := make(map[string]int)

	result := &PredictionResult{
		CurrentAnte: currentAnte,
		Seed:        seed,
		Antes:       make([]AntePrediction, 0, maxAnte),
		NextTags:    make([]string, 0, 3),
	}

	// 1. Predict Ante progressions
	for a := 1; a <= maxAnte; a++ {
		antePred := AntePrediction{
			Ante: a,
		}

		// Boss Blind prediction
		antePred.Boss, antePred.BossKey = predictBoss(rng, a, bossesUsed)

		// Voucher prediction for this Ante
		antePred.Voucher = predictVoucher(rng, a, usedVouchers)

		// If this ante has already passed or is current, consider its voucher as taken for future projections
		if a < currentAnte || usedVouchers[antePred.Voucher.Key] {
			usedVouchers[antePred.Voucher.Key] = true
		}

		result.Antes = append(result.Antes, antePred)
	}

	// 2. Predict upcoming Voucher Tags queue (next 3)
	tagRNG := rng.Clone()
	tagPoolVouchers := make(map[string]bool, len(usedVouchers))
	for k, v := range usedVouchers {
		tagPoolVouchers[k] = v
	}

	for i := 1; i <= 3; i++ {
		tagV := predictVoucherFromTag(tagRNG, tagPoolVouchers)
		result.NextTags = append(result.NextTags, tagV.Name)
		tagPoolVouchers[tagV.Key] = true
	}

	return result
}

func predictBoss(rng *BalatroRNG, ante int, bossesUsed map[string]int) (string, string) {
	// Build eligible bosses according to Balatro rules
	eligible := make([]BlindDef, 0, len(MasterBossBlinds))
	minUse := 999

	isShowdown := ante >= 2 && ante%8 == 0

	for _, b := range MasterBossBlinds {
		if isShowdown {
			if b.Showdown {
				eligible = append(eligible, b)
				count := bossesUsed[b.Key]
				if count < minUse {
					minUse = count
				}
			}
		} else {
			if !b.Showdown && b.MinAnte <= ante {
				eligible = append(eligible, b)
				count := bossesUsed[b.Key]
				if count < minUse {
					minUse = count
				}
			}
		}
	}

	// Filter down to min_use pool
	filtered := make([]string, 0, len(eligible))
	keyToName := make(map[string]string, len(eligible))
	for _, b := range eligible {
		if bossesUsed[b.Key] <= minUse {
			filtered = append(filtered, b.Key)
			keyToName[b.Key] = b.Name
		}
	}

	if len(filtered) == 0 {
		return "The Hook", "bl_hook"
	}

	streamKey := fmt.Sprintf("boss_%d", ante)
	chosenKey, _ := rng.PseudorandomElement(filtered, streamKey)
	bossesUsed[chosenKey]++

	return keyToName[chosenKey], chosenKey
}

func predictVoucher(rng *BalatroRNG, ante int, used map[string]bool) VoucherPred {
	pool := make([]string, len(MasterVouchers))
	for i, v := range MasterVouchers {
		if v.Requires == "" {
			if used[v.Key] {
				pool[i] = "UNAVAILABLE"
			} else {
				pool[i] = v.Key
			}
		} else {
			if used[v.Requires] && !used[v.Key] {
				pool[i] = v.Key
			} else {
				pool[i] = "UNAVAILABLE"
			}
		}
	}

	streamKey := fmt.Sprintf("Voucher%d", ante)
	choice, rawIdx := rng.PseudorandomElement(pool, streamKey)
	it := 1
	for choice == "UNAVAILABLE" {
		it++
		resampleKey := fmt.Sprintf("%s_resample%d", streamKey, it)
		choice, rawIdx = rng.PseudorandomElement(pool, resampleKey)
	}

	vDef := MasterVouchers[rawIdx]
	return VoucherPred{
		Key:      vDef.Key,
		Name:     vDef.Name,
		PosX:     vDef.PosX,
		PosY:     vDef.PosY,
		Resample: it,
	}
}

func predictVoucherFromTag(rng *BalatroRNG, used map[string]bool) VoucherPred {
	pool := make([]string, len(MasterVouchers))
	for i, v := range MasterVouchers {
		if v.Requires == "" {
			if used[v.Key] {
				pool[i] = "UNAVAILABLE"
			} else {
				pool[i] = v.Key
			}
		} else {
			if used[v.Requires] && !used[v.Key] {
				pool[i] = v.Key
			} else {
				pool[i] = "UNAVAILABLE"
			}
		}
	}

	streamKey := "Voucher_fromtag"
	choice, rawIdx := rng.PseudorandomElement(pool, streamKey)
	it := 1
	for choice == "UNAVAILABLE" {
		it++
		resampleKey := fmt.Sprintf("%s_resample%d", streamKey, it)
		choice, rawIdx = rng.PseudorandomElement(pool, resampleKey)
	}

	vDef := MasterVouchers[rawIdx]
	return VoucherPred{
		Key:      vDef.Key,
		Name:     vDef.Name,
		PosX:     vDef.PosX,
		PosY:     vDef.PosY,
		Resample: it,
	}
}
