package analyzer

import (
	"fmt"
	"strings"
)

// PackMeta defines pack properties.
type PackMeta struct {
	Type    string
	Size    int
	Choices int
	Title   string
}

// GetPackMeta extracts type, size, choices, and friendly title for a pack.
func GetPackMeta(packName string) PackMeta {
	switch packName {
	case "Arcana Pack":
		return PackMeta{Type: "Arcana Pack", Size: 3, Choices: 1, Title: "Arcana 3 cards, pick 1"}
	case "Jumbo Arcana Pack":
		return PackMeta{Type: "Arcana Pack", Size: 5, Choices: 1, Title: "Arcana 5 cards, pick 1"}
	case "Mega Arcana Pack":
		return PackMeta{Type: "Arcana Pack", Size: 5, Choices: 2, Title: "Arcana 5 cards, pick 2"}

	case "Celestial Pack":
		return PackMeta{Type: "Celestial Pack", Size: 3, Choices: 1, Title: "Celestial 3 cards, pick 1"}
	case "Jumbo Celestial Pack":
		return PackMeta{Type: "Celestial Pack", Size: 5, Choices: 1, Title: "Celestial 5 cards, pick 1"}
	case "Mega Celestial Pack":
		return PackMeta{Type: "Celestial Pack", Size: 5, Choices: 1, Title: "Celestial 5 cards, pick 1"}

	case "Buffoon Pack":
		return PackMeta{Type: "Buffoon Pack", Size: 2, Choices: 1, Title: "Buffoon 2 cards, pick 1"}
	case "Jumbo Buffoon Pack":
		return PackMeta{Type: "Buffoon Pack", Size: 4, Choices: 1, Title: "Buffoon 4 cards, pick 1"}
	case "Mega Buffoon Pack":
		return PackMeta{Type: "Buffoon Pack", Size: 4, Choices: 2, Title: "Buffoon 4 cards, pick 2"}

	case "Spectral Pack":
		return PackMeta{Type: "Spectral Pack", Size: 2, Choices: 1, Title: "Spectral 2 cards, pick 1"}
	case "Jumbo Spectral Pack":
		return PackMeta{Type: "Spectral Pack", Size: 4, Choices: 1, Title: "Spectral 4 cards, pick 1"}
	case "Mega Spectral Pack":
		return PackMeta{Type: "Spectral Pack", Size: 4, Choices: 1, Title: "Spectral 4 cards, pick 1"}

	case "Standard Pack":
		return PackMeta{Type: "Standard Pack", Size: 3, Choices: 1, Title: "Standard 3 cards, pick 1"}
	case "Jumbo Standard Pack":
		return PackMeta{Type: "Standard Pack", Size: 5, Choices: 1, Title: "Standard 5 cards, pick 1"}
	case "Mega Standard Pack":
		return PackMeta{Type: "Standard Pack", Size: 5, Choices: 1, Title: "Standard 5 cards, pick 1"}

	default:
		return PackMeta{Type: packName, Size: 3, Choices: 1, Title: packName}
	}
}

// NextBoss determines the boss blind for an ante, respecting showdown ante multiples.
func (inst *Instance) NextBoss(ante int) string {
	var bossPool []string
	isShowdown := (ante%8 == 0)

	for _, b := range BOSSES {
		if !inst.IsLocked(b) {
			isBossShowdown := (b[0] != 'T') // Finisher showdown bosses: Amber Acorn, Cerulean Bell, Crimson Heart, Verdant Leaf, Violet Vessel
			if (isShowdown && isBossShowdown) || (!isShowdown && !isBossShowdown) {
				bossPool = append(bossPool, b)
			}
		}
	}

	if len(bossPool) == 0 {
		for _, b := range BOSSES {
			isBossShowdown := (b[0] != 'T')
			if (isShowdown && isBossShowdown) || (!isShowdown && !isBossShowdown) {
				inst.Unlock(b)
			}
		}
		return inst.NextBoss(ante)
	}

	chosen := inst.RandChoice("boss", bossPool)
	inst.Lock(chosen)
	return chosen
}

// NextVoucher determines the voucher for an ante.
func (inst *Instance) NextVoucher(ante int) string {
	return inst.RandChoice(fmt.Sprintf("Voucher%d", ante), VOUCHERS)
}

// NextVoucherFromTag determines the voucher created by skipping on a Voucher Tag.
func (inst *Instance) NextVoucherFromTag() string {
	return inst.RandChoice("Voucher_fromtag", VOUCHERS)
}

// NextTag determines a skip tag for an ante.
func (inst *Instance) NextTag(ante int) string {
	return inst.RandChoice(fmt.Sprintf("Tag%d", ante), TAGS)
}

// NextPack determines the booster pack rolled in shop.
func (inst *Instance) NextPack(ante int) string {
	if ante <= 2 && !inst.GeneratedFirstPack {
		inst.GeneratedFirstPack = true
		return "Buffoon Pack"
	}
	return inst.RandWeightedChoice(fmt.Sprintf("shop_pack%d", ante), PACKS)
}

// NextJoker generates a Joker with edition and stake stickers.
func (inst *Instance) NextJoker(source string, ante int, hasStickers bool) CardItem {
	anteStr := fmt.Sprintf("%d", ante)

	// 1. Rarity
	rarity := "1"
	if source == "sou" {
		rarity = "4"
	} else if source == "wra" || source == "rta" {
		rarity = "3"
	} else if source == "uta" {
		rarity = "2"
	} else if source == "top" {
		rarity = "1"
	} else {
		poll := inst.Random(fmt.Sprintf("rarity%s%s", anteStr, source))
		if poll > 0.95 {
			rarity = "3"
		} else if poll > 0.70 {
			rarity = "2"
		} else {
			rarity = "1"
		}
	}

	// 2. Edition
	editionRate := 1
	if inst.IsVoucherActive("Glow Up") {
		editionRate = 4
	} else if inst.IsVoucherActive("Hone") {
		editionRate = 2
	}

	edition := ""
	ediPoll := inst.Random(fmt.Sprintf("edi%s%s", source, anteStr))
	if ediPoll > 0.997 {
		edition = "Negative"
	} else if ediPoll > 1.0-0.006*float64(editionRate) {
		edition = "Polychrome"
	} else if ediPoll > 1.0-0.02*float64(editionRate) {
		edition = "Holographic"
	} else if ediPoll > 1.0-0.04*float64(editionRate) {
		edition = "Foil"
	}

	// 3. Joker Name
	var jokerName string
	switch rarity {
	case "4":
		jokerName = inst.RandChoice("Joker4", LEGENDARY_JOKERS)
	case "3":
		jokerName = inst.RandChoice(fmt.Sprintf("Joker3%s%s", source, anteStr), RARE_JOKERS)
	case "2":
		jokerName = inst.RandChoice(fmt.Sprintf("Joker2%s%s", source, anteStr), UNCOMMON_JOKERS)
	default:
		jokerName = inst.RandChoice(fmt.Sprintf("Joker1%s%s", source, anteStr), COMMON_JOKERS)
	}

	// 4. Stickers (Eternal, Perishable, Rental)
	var stickers CardStickers
	if hasStickers {
		prefix := "etperpoll"
		if source == "buf" {
			prefix = "packetper"
		}
		stickerPoll := inst.Random(prefix + anteStr)

		// Eternal on Black Stake (3) and above
		if stickerPoll > 0.7 && inst.Params.Stake >= 3 {
			if !isEternalExempt(jokerName) {
				stickers.Eternal = true
			}
		}

		// Perishable on Orange Stake (5) and above
		if (stickerPoll > 0.4 && stickerPoll <= 0.7) && inst.Params.Stake >= 5 {
			if !isPerishableExempt(jokerName) {
				stickers.Perishable = true
			}
		}

		// Rental on Gold Stake (6) and above
		if inst.Params.Stake >= 6 {
			rPrefix := "ssjr"
			if source == "buf" {
				rPrefix = "packssjr"
			}
			if inst.Random(rPrefix+anteStr) > 0.7 {
				stickers.Rental = true
			}
		}
	}

	pos := JOKER_POS[jokerName]
	return CardItem{
		Name:     jokerName,
		Set:      "Joker",
		PosX:     pos.X,
		PosY:     pos.Y,
		Edition:  edition,
		Stickers: stickers,
	}
}

func isEternalExempt(name string) bool {
	switch name {
	case "Gros Michel", "Ice Cream", "Cavendish", "Luchador", "Turtle Bean",
		"Diet Cola", "Popcorn", "Ramen", "Seltzer", "Mr. Bones", "Invisible Joker":
		return true
	}
	return false
}

func isPerishableExempt(name string) bool {
	switch name {
	case "Ceremonial Dagger", "Ride the Bus", "Runner", "Constellation",
		"Green Joker", "Red Card", "Madness", "Square Joker",
		"Vampire", "Rocket", "Obelisk", "Lucky Cat",
		"Flash Card", "Spare Trousers", "Castle", "Wee Joker":
		return true
	}
	return false
}

// NextTarot generates a Tarot card with 0.3% chance of The Soul.
func (inst *Instance) NextTarot(source string, ante int, soulable bool) CardItem {
	anteStr := fmt.Sprintf("%d", ante)
	if soulable && (inst.Params.Showman || !inst.IsLocked("The Soul")) {
		if inst.Random("soul_Tarot"+anteStr) > 0.997 {
			pos := TAROT_POS["The Soul"]
			return CardItem{Name: "The Soul", Set: "Tarot", PosX: pos.X, PosY: pos.Y}
		}
	}
	name := inst.RandChoice(fmt.Sprintf("Tarot%s%s", source, anteStr), TAROTS)
	pos := TAROT_POS[name]
	return CardItem{Name: name, Set: "Tarot", PosX: pos.X, PosY: pos.Y}
}

// NextPlanet generates a Planet card with 0.3% chance of Black Hole.
func (inst *Instance) NextPlanet(source string, ante int, soulable bool) CardItem {
	anteStr := fmt.Sprintf("%d", ante)
	if soulable && (inst.Params.Showman || !inst.IsLocked("Black Hole")) {
		if inst.Random("soul_Planet"+anteStr) > 0.997 {
			pos := TAROT_POS["Black Hole"]
			return CardItem{Name: "Black Hole", Set: "Planet", PosX: pos.X, PosY: pos.Y}
		}
	}
	name := inst.RandChoice(fmt.Sprintf("Planet%s%s", source, anteStr), PLANETS)
	pos := TAROT_POS[name]
	return CardItem{Name: name, Set: "Planet", PosX: pos.X, PosY: pos.Y}
}

// NextSpectral generates a Spectral card with soulable check.
func (inst *Instance) NextSpectral(source string, ante int, soulable bool) CardItem {
	anteStr := fmt.Sprintf("%d", ante)
	forcedKey := ""
	if soulable {
		if inst.Params.Showman || !inst.IsLocked("The Soul") {
			if inst.Random("soul_Spectral"+anteStr) > 0.997 {
				forcedKey = "The Soul"
			}
		}
		if inst.Params.Showman || !inst.IsLocked("Black Hole") {
			if inst.Random("soul_Spectral"+anteStr) > 0.997 {
				forcedKey = "Black Hole"
			}
		}
	}
	if forcedKey != "" {
		pos := TAROT_POS[forcedKey]
		return CardItem{Name: forcedKey, Set: "Spectral", PosX: pos.X, PosY: pos.Y}
	}
	name := inst.RandChoice(fmt.Sprintf("Spectral%s%s", source, anteStr), SPECTRALS)
	pos := TAROT_POS[name]
	return CardItem{Name: name, Set: "Spectral", PosX: pos.X, PosY: pos.Y}
}

// NextStandardCard generates a playing card with base suit/rank, enhancement, edition, and seal.
func (inst *Instance) NextStandardCard(ante int) CardItem {
	anteStr := fmt.Sprintf("%d", ante)

	// Enhancement
	enhancement := ""
	if inst.Random("stdset"+anteStr) > 0.6 {
		enhancement = inst.RandChoice("Enhancedsta"+anteStr, ENHANCEMENTS)
	}

	// Base card
	baseCode := inst.RandChoice("frontsta"+anteStr, CARDS)
	suitName := "Spades"
	switch baseCode[0] {
	case 'C':
		suitName = "Clubs"
	case 'D':
		suitName = "Diamonds"
	case 'H':
		suitName = "Hearts"
	case 'S':
		suitName = "Spades"
	}

	rankName := string(baseCode[2])
	switch rankName {
	case "T":
		rankName = "10"
	case "J":
		rankName = "Jack"
	case "Q":
		rankName = "Queen"
	case "K":
		rankName = "King"
	case "A":
		rankName = "Ace"
	}

	// Edition
	edition := ""
	edPoll := inst.Random("standard_edition" + anteStr)
	if edPoll > 0.988 {
		edition = "Polychrome"
	} else if edPoll > 0.96 {
		edition = "Holographic"
	} else if edPoll > 0.92 {
		edition = "Foil"
	}

	// Seal
	seal := ""
	if inst.Random("stdseal"+anteStr) > 0.8 {
		sPoll := inst.Random("stdsealtype" + anteStr)
		if sPoll > 0.75 {
			seal = "Red"
		} else if sPoll > 0.50 {
			seal = "Blue"
		} else if sPoll > 0.25 {
			seal = "Gold"
		} else {
			seal = "Purple"
		}
	}

	displayName := fmt.Sprintf("%s of %s", rankName, suitName)
	if enhancement != "" {
		displayName = fmt.Sprintf("%s (%s)", displayName, enhancement)
	}

	return CardItem{
		Name:        displayName,
		Set:         "Playing Card",
		Rank:        rankName,
		Suit:        suitName,
		Enhancement: enhancement,
		Edition:     edition,
		Seal:        seal,
	}
}

// NextBuffoonPack unrolls Jokers for a Buffoon pack without offering internal duplicates.
func (inst *Instance) NextBuffoonPack(size int, ante int) []CardItem {
	cards := make([]CardItem, 0, size)
	lockedByPack := make([]string, 0, size)
	for i := 0; i < size; i++ {
		jkr := inst.NextJoker("buf", ante, true)
		cards = append(cards, jkr)
		if !inst.Params.Showman && !inst.IsLocked(jkr.Name) {
			inst.Lock(jkr.Name)
			lockedByPack = append(lockedByPack, jkr.Name)
		}
	}
	for _, name := range lockedByPack {
		inst.Unlock(name)
	}
	return cards
}

// NextArcanaPack unrolls Tarots (or Spectrals with Omen Globe) for an Arcana pack.
func (inst *Instance) NextArcanaPack(size int, ante int) []CardItem {
	cards := make([]CardItem, 0, size)
	lockedByPack := make([]string, 0, size)
	for i := 0; i < size; i++ {
		var card CardItem
		if inst.IsVoucherActive("Omen Globe") && inst.Random("omen_globe") > 0.8 {
			card = inst.NextSpectral("ar2", ante, true)
		} else {
			card = inst.NextTarot("ar1", ante, true)
		}
		cards = append(cards, card)
		if !inst.Params.Showman && !inst.IsLocked(card.Name) {
			inst.Lock(card.Name)
			lockedByPack = append(lockedByPack, card.Name)
		}
	}
	for _, name := range lockedByPack {
		inst.Unlock(name)
	}
	return cards
}

// NextCelestialPack unrolls Planet cards for a Celestial pack.
func (inst *Instance) NextCelestialPack(size int, ante int) []CardItem {
	cards := make([]CardItem, 0, size)
	lockedByPack := make([]string, 0, size)
	for i := 0; i < size; i++ {
		card := inst.NextPlanet("pl1", ante, true)
		cards = append(cards, card)
		if !inst.Params.Showman && !inst.IsLocked(card.Name) {
			inst.Lock(card.Name)
			lockedByPack = append(lockedByPack, card.Name)
		}
	}
	for _, name := range lockedByPack {
		inst.Unlock(name)
	}
	return cards
}

// NextSpectralPack unrolls Spectral cards for a Spectral pack.
func (inst *Instance) NextSpectralPack(size int, ante int) []CardItem {
	cards := make([]CardItem, 0, size)
	lockedByPack := make([]string, 0, size)
	for i := 0; i < size; i++ {
		card := inst.NextSpectral("spe", ante, true)
		cards = append(cards, card)
		if !inst.Params.Showman && !inst.IsLocked(card.Name) {
			inst.Lock(card.Name)
			lockedByPack = append(lockedByPack, card.Name)
		}
	}
	for _, name := range lockedByPack {
		inst.Unlock(name)
	}
	return cards
}

// NextStandardPack unrolls playing cards for a Standard pack.
func (inst *Instance) NextStandardPack(size int, ante int) []CardItem {
	cards := make([]CardItem, 0, size)
	for i := 0; i < size; i++ {
		cards = append(cards, inst.NextStandardCard(ante))
	}
	return cards
}

// AnalyzeSeed executes high-accuracy Balatro seed prediction.
func AnalyzeSeed(
	seed string,
	currentAnte int,
	redeemedVouchers []string,
	ownedJokers []string,
	ownedConsumables []string,
	poolFlags map[string]bool,
	deck string,
	stake int,
	maxAnte int,
) *PredictionResult {
	clean := CleanSeed(seed)
	if clean == "" {
		return nil
	}
	if maxAnte <= 0 {
		maxAnte = 16
	}
	if currentAnte+2 > maxAnte {
		maxAnte = currentAnte + 2
	}

	inst := NewInstance(clean, deck, stake, poolFlags)
	inst.InitLocks(1, false, true)

	// Check if Showman is owned
	for _, j := range ownedJokers {
		if j == "Showman" {
			inst.Params.Showman = true
			break
		}
	}

	// Apply already redeemed vouchers
	for _, v := range redeemedVouchers {
		inst.ActivateVoucher(v)
	}

	// Lock currently held jokers (culls them from pools unless Showman)
	if !inst.Params.Showman {
		for _, j := range ownedJokers {
			inst.Lock(j)
		}
	}

	// Lock currently held consumables (Pluto, Death, etc. cull from pools unless Showman)
	if !inst.Params.Showman {
		for _, c := range ownedConsumables {
			inst.Lock(c)
		}
	}

	result := &PredictionResult{
		CurrentAnte: currentAnte,
		Seed:        clean,
		Deck:        deck,
		Stake:       stake,
		Antes:       make([]AntePrediction, 0, maxAnte),
	}

	for a := 1; a <= maxAnte; a++ {
		inst.InitUnlocks(a, false)

		antePred := AntePrediction{
			Ante:  a,
			Tags:  make([]TagPrediction, 0, 2),
			Packs: make([]BlindPacks, 0, 3),
		}

		// 1. Boss Blind prediction + Director's Cut & Retcon rerolls
		bossName := inst.NextBoss(a)
		bossPos := BOSS_POS[bossName]

		bossPred := BossPrediction{
			Name: bossName,
			Key:  strings.ToLower(strings.ReplaceAll(bossName, " ", "_")),
			PosX: bossPos.X,
			PosY: bossPos.Y,
		}

		// Simulate rerolls on a clone so main campaign stream is not consumed
		rerollClone := inst.Clone()
		r1 := rerollClone.NextBoss(a)
		r1Pos := BOSS_POS[r1]
		bossPred.DirectorCutReroll = &BossRef{
			Name: r1,
			Key:  strings.ToLower(strings.ReplaceAll(r1, " ", "_")),
			PosX: r1Pos.X,
			PosY: r1Pos.Y,
		}

		r2 := rerollClone.NextBoss(a)
		r2Pos := BOSS_POS[r2]
		r3 := rerollClone.NextBoss(a)
		r3Pos := BOSS_POS[r3]
		bossPred.RetconRerolls = []BossRef{
			{Name: r2, Key: strings.ToLower(strings.ReplaceAll(r2, " ", "_")), PosX: r2Pos.X, PosY: r2Pos.Y},
			{Name: r3, Key: strings.ToLower(strings.ReplaceAll(r3, " ", "_")), PosX: r3Pos.X, PosY: r3Pos.Y},
		}
		antePred.Boss = bossPred

		// 2. Voucher prediction for this Ante
		vName := inst.NextVoucher(a)
		vPos := VOUCHER_POS[vName]
		antePred.Voucher = VoucherPrediction{
			Key:  vName,
			Name: vName,
			PosX: vPos.X,
			PosY: vPos.Y,
		}

		// 3. Tags prediction (Small Blind and Big Blind)
		smallTag := inst.NextTag(a)
		bigTag := inst.NextTag(a)

		antePred.Tags = append(antePred.Tags, predictTagReward(inst, smallTag, "Small Blind", a))
		antePred.Tags = append(antePred.Tags, predictTagReward(inst, bigTag, "Big Blind", a))

		// 4. Packs prediction
		// Ante 1: 4 packs total (0 in Small Blind, 2 in Big Blind, 2 in Boss Blind)
		// Ante 2+: 6 packs total (2 in Small Blind, 2 in Big Blind, 2 in Boss Blind)
		if a == 1 {
			// Small Blind: none
			antePred.Packs = append(antePred.Packs, BlindPacks{
				Blind: "Small Blind",
				Packs: []PackPrediction{},
			})

			// Big Blind: 2 packs
			bbPacks := make([]PackPrediction, 0, 2)
			for p := 0; p < 2; p++ {
				bbPacks = append(bbPacks, generatePackPrediction(inst, a))
			}
			antePred.Packs = append(antePred.Packs, BlindPacks{
				Blind: "Big Blind",
				Packs: bbPacks,
			})

			// Boss Blind: 2 packs
			bossPacks := make([]PackPrediction, 0, 2)
			for p := 0; p < 2; p++ {
				bossPacks = append(bossPacks, generatePackPrediction(inst, a))
			}
			antePred.Packs = append(antePred.Packs, BlindPacks{
				Blind: "Boss Blind",
				Packs: bossPacks,
			})
		} else {
			// Small Blind: 2 packs
			sbPacks := make([]PackPrediction, 0, 2)
			for p := 0; p < 2; p++ {
				sbPacks = append(sbPacks, generatePackPrediction(inst, a))
			}
			antePred.Packs = append(antePred.Packs, BlindPacks{
				Blind: "Small Blind",
				Packs: sbPacks,
			})

			// Big Blind: 2 packs
			bbPacks := make([]PackPrediction, 0, 2)
			for p := 0; p < 2; p++ {
				bbPacks = append(bbPacks, generatePackPrediction(inst, a))
			}
			antePred.Packs = append(antePred.Packs, BlindPacks{
				Blind: "Big Blind",
				Packs: bbPacks,
			})

			// Boss Blind: 2 packs
			bossPacks := make([]PackPrediction, 0, 2)
			for p := 0; p < 2; p++ {
				bossPacks = append(bossPacks, generatePackPrediction(inst, a))
			}
			antePred.Packs = append(antePred.Packs, BlindPacks{
				Blind: "Boss Blind",
				Packs: bossPacks,
			})
		}

		result.Antes = append(result.Antes, antePred)
	}

	return result
}

// generatePackPrediction generates a booster pack and resolves its inner cards.
func generatePackPrediction(inst *Instance, ante int) PackPrediction {
	packName := inst.NextPack(ante)
	meta := GetPackMeta(packName)

	pred := PackPrediction{
		Name:    packName,
		Type:    meta.Type,
		Size:    meta.Size,
		Choices: meta.Choices,
		Title:   meta.Title,
	}

	switch meta.Type {
	case "Buffoon Pack":
		pred.Cards = inst.NextBuffoonPack(meta.Size, ante)
	case "Arcana Pack":
		pred.Cards = inst.NextArcanaPack(meta.Size, ante)
	case "Celestial Pack":
		pred.Cards = inst.NextCelestialPack(meta.Size, ante)
	case "Spectral Pack":
		pred.Cards = inst.NextSpectralPack(meta.Size, ante)
	case "Standard Pack":
		pred.Cards = inst.NextStandardPack(meta.Size, ante)
	}

	return pred
}

// predictTagReward predicts what skipping on a specific tag yields on a cloned branch.
func predictTagReward(inst *Instance, tagName string, blind string, ante int) TagPrediction {
	pos := TAG_POS[tagName]
	pred := TagPrediction{
		Name:  tagName,
		Blind: blind,
		PosX:  pos.X,
		PosY:  pos.Y,
	}

	clone := inst.Clone()

	switch tagName {
	case "Voucher Tag":
		pred.RewardType = "voucher"
		v := clone.NextVoucherFromTag()
		vPos := VOUCHER_POS[v]
		pred.RewardTitle = fmt.Sprintf("Voucher: %s", v)
		pred.RewardVoucher = &VoucherPrediction{
			Key:  v,
			Name: v,
			PosX: vPos.X,
			PosY: vPos.Y,
		}

	case "Charm Tag":
		pred.RewardType = "pack"
		pred.RewardTitle = "Free Mega Arcana Pack (5 cards, pick 2)"
		pred.RewardCards = clone.NextArcanaPack(5, ante)

	case "Meteor Tag":
		pred.RewardType = "pack"
		pred.RewardTitle = "Free Mega Celestial Pack (5 cards, pick 1)"
		pred.RewardCards = clone.NextCelestialPack(5, ante)

	case "Buffoon Tag":
		pred.RewardType = "pack"
		pred.RewardTitle = "Free Mega Buffoon Pack (4 cards, pick 2)"
		pred.RewardCards = clone.NextBuffoonPack(4, ante)

	case "Ethereal Tag":
		pred.RewardType = "pack"
		pred.RewardTitle = "Free Spectral Pack (4 cards, pick 1)"
		pred.RewardCards = clone.NextSpectralPack(4, ante)

	case "Standard Tag":
		pred.RewardType = "pack"
		pred.RewardTitle = "Free Mega Standard Pack (5 cards, pick 1)"
		pred.RewardCards = clone.NextStandardPack(5, ante)

	case "Uncommon Tag":
		pred.RewardType = "joker"
		j := clone.NextJoker("uta", ante, false)
		pred.RewardTitle = fmt.Sprintf("Free Uncommon Joker: %s", j.Name)
		pred.RewardCards = []CardItem{j}

	case "Rare Tag":
		pred.RewardType = "joker"
		j := clone.NextJoker("rta", ante, false)
		pred.RewardTitle = fmt.Sprintf("Free Rare Joker: %s", j.Name)
		pred.RewardCards = []CardItem{j}

	case "Top-up Tag":
		pred.RewardType = "joker"
		pred.RewardTitle = "Spawns 2 Common Jokers"
		pred.RewardCards = []CardItem{
			clone.NextJoker("top", ante, false),
			clone.NextJoker("top", ante, false),
		}

	case "Orbital Tag":
		pred.RewardType = "hand"
		hand := clone.RandChoice("orbital", POKER_HANDS)
		pred.RewardHand = hand
		pred.RewardTitle = fmt.Sprintf("+3 levels to %s", hand)

	case "Coupon Tag":
		pred.RewardType = "bonus"
		pred.RewardTitle = "Initial cards and booster packs in next shop are free"

	case "D6 Tag":
		pred.RewardType = "bonus"
		pred.RewardTitle = "Next shop reroll starts at $0"

	case "Foil Tag":
		pred.RewardType = "bonus"
		pred.RewardTitle = "Next base edition shop Joker is Foil (+$50 Chips)"

	case "Holographic Tag":
		pred.RewardType = "bonus"
		pred.RewardTitle = "Next base edition shop Joker is Holographic (+10 Mult)"

	case "Polychrome Tag":
		pred.RewardType = "bonus"
		pred.RewardTitle = "Next base edition shop Joker is Polychrome (1.5x Mult)"

	case "Investment Tag":
		pred.RewardType = "bonus"
		pred.RewardTitle = "Defeating the Boss Blind awards $25"

	case "Economy Tag":
		pred.RewardType = "bonus"
		pred.RewardTitle = "Doubles your current money (up to +$40)"

	case "Speed Tag":
		pred.RewardType = "bonus"
		pred.RewardTitle = "Gives $5 per skipped blind this run"

	default:
		pred.RewardType = "none"
		pred.RewardTitle = tagName
	}

	return pred
}
