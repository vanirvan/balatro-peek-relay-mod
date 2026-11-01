package analyzer

import (
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"
	"testing"
)

func TestPseudohash(t *testing.T) {
	val := Pseudohash("IMMOLATE")
	if val <= 0.0 || val >= 1.0 {
		t.Fatalf("expected 0 < val < 1, got %f", val)
	}
}

func TestLuaRandom(t *testing.T) {
	rng := NewLuaRandom(0.5)
	r := rng.Random()
	if r < 0.0 || r >= 1.0 {
		t.Fatalf("expected 0 <= r < 1, got %f", r)
	}
}

func TestAnalyzeSeed(t *testing.T) {
	pred := AnalyzeSeed("IMMOLATE", 1, nil, nil, nil, nil, "Red Deck", 1, 16)
	if pred == nil {
		t.Fatal("expected non-nil prediction")
	}
	if len(pred.Antes) < 16 {
		t.Fatalf("expected at least 16 antes, got %d", len(pred.Antes))
	}

	ante1 := pred.Antes[0]
	t.Logf("Ante 1 Boss: %s (Reroll 1: %s, Reroll 2: %s)", ante1.Boss.Name, ante1.Boss.DirectorCutReroll.Name, ante1.Boss.RetconRerolls[0].Name)
	t.Logf("Ante 1 Voucher: %s", ante1.Voucher.Name)
	t.Logf("Ante 1 Tags: %s, %s", ante1.Tags[0].Name, ante1.Tags[1].Name)
	t.Logf("Ante 1 Packs count: %d", len(ante1.Packs))
	for _, bp := range ante1.Packs {
		for _, p := range bp.Packs {
			t.Logf("  [%s] Pack: %s (%s, %d cards)", bp.Blind, p.Name, p.Title, len(p.Cards))
		}
	}
}

func TestSeed53841CZI(t *testing.T) {
	// Test Ante 3 holding Pluto
	pred := AnalyzeSeed("53841CZI", 3, nil, nil, []string{"Pluto"}, nil, "Red Deck", 1, 3)
	ante3 := pred.Antes[2]
	for _, bp := range ante3.Packs {
		for _, p := range bp.Packs {
			t.Logf("[%s] Pack: %s (%s)", bp.Blind, p.Name, p.Title)
			for _, c := range p.Cards {
				t.Logf("  Card: %s", c.Name)
			}
		}
	}

	// Verify Small Blind Celestial Pack with Pluto held:
	// Should produce: Venus, Neptune, Mercury
	sbPack := ante3.Packs[0].Packs[1]
	if sbPack.Cards[0].Name != "Venus" || sbPack.Cards[1].Name != "Neptune" || sbPack.Cards[2].Name != "Mercury" {
		t.Fatalf("expected Small Blind pack to be Venus, Neptune, Mercury; got %s, %s, %s",
			sbPack.Cards[0].Name, sbPack.Cards[1].Name, sbPack.Cards[2].Name)
	}

	// Verify Big Blind Jumbo Celestial Pack:
	// Should produce: Mars, Venus, Saturn, Earth, Uranus
	bbPack := ante3.Packs[1].Packs[1]
	if bbPack.Cards[4].Name != "Uranus" {
		t.Fatalf("expected Big Blind card 5 to be Uranus; got %s", bbPack.Cards[4].Name)
	}

	// Verify Boss Blind Celestial Pack:
	// Should produce: Mercury, Saturn, ...
	bossPack := ante3.Packs[2].Packs[0]
	if bossPack.Cards[0].Name != "Mercury" || bossPack.Cards[1].Name != "Saturn" {
		t.Fatalf("expected Boss Blind pack cards 1 & 2 to be Mercury, Saturn; got %s, %s",
			bossPack.Cards[0].Name, bossPack.Cards[1].Name)
	}
}

func TestTraceAnte3WithPlutoHeld(t *testing.T) {
	inst := NewInstance("53841CZI", "Red Deck", 1, nil)
	inst.InitLocks(1, false, true)

	// Ante 1
	_ = inst.NextBoss(1)
	_ = inst.NextVoucher(1)
	_ = inst.NextTag(1)
	_ = inst.NextTag(1)
	_ = inst.NextBuffoonPack(2, 1)
	_ = inst.NextArcanaPack(3, 1)
	_ = inst.NextArcanaPack(5, 1)
	_ = inst.NextArcanaPack(3, 1)

	// Ante 2
	inst.InitUnlocks(2, false)
	_ = inst.NextBoss(2)
	_ = inst.NextVoucher(2)
	_ = inst.NextTag(2)
	_ = inst.NextTag(2)
	_ = inst.NextStandardPack(5, 2)
	_ = inst.NextBuffoonPack(4, 2)
	_ = inst.NextArcanaPack(3, 2)
	_ = inst.NextSpectralPack(4, 2)
	_ = inst.NextCelestialPack(3, 2)
	_ = inst.NextCelestialPack(3, 2)

	// Ante 3
	inst.InitUnlocks(3, false)
	_ = inst.NextBoss(3)
	_ = inst.NextVoucher(3)
	_ = inst.NextTag(3)
	_ = inst.NextTag(3)

	_ = inst.NextArcanaPack(5, 3)

	// In Small Blind: PLAYER HOLDS PLUTO!
	inst.Lock("Pluto")
	t.Logf("--- Ante 3 Small Blind Celestial Pack (HOLDING PLUTO) ---")
	sbCel := inst.NextCelestialPack(3, 3)
	for i, c := range sbCel {
		t.Logf("  Card %d: %s", i+1, c.Name)
	}

	_ = inst.NextStandardPack(3, 3)

	t.Logf("--- Ante 3 Big Blind Jumbo Celestial Pack (STILL HOLDING PLUTO) ---")
	bbCel := inst.NextCelestialPack(5, 3)
	for i, c := range bbCel {
		t.Logf("  Card %d: %s", i+1, c.Name)
	}

	t.Logf("--- Ante 3 Boss Blind Celestial Pack (STILL HOLDING PLUTO) ---")
	bossCel := inst.NextCelestialPack(3, 3)
	for i, c := range bossCel {
		t.Logf("  Card %d: %s", i+1, c.Name)
	}
}

func TestSeedXwv7eaip(t *testing.T) {
	pred := AnalyzeSeed("xwv7eaip", 3, nil, nil, nil, nil, "Red Deck", 1, 3)
	ante3 := pred.Antes[2]
	bbPacks := ante3.Packs[1].Packs
	var specPack *PackPrediction
	for i := range bbPacks {
		if bbPacks[i].Name == "Spectral Pack" {
			specPack = &bbPacks[i]
			break
		}
	}
	if specPack == nil {
		t.Fatal("expected Spectral Pack in Ante 3 Big Blind")
	}
	if len(specPack.Cards) != 2 {
		t.Fatalf("expected 2 cards in Spectral Pack, got %d", len(specPack.Cards))
	}
	t.Logf("Spectral Pack cards: %s, %s", specPack.Cards[0].Name, specPack.Cards[1].Name)
	if specPack.Cards[0].Name != "Familiar" || specPack.Cards[1].Name != "Medium" {
		t.Fatalf("expected Familiar and Medium; got %s and %s", specPack.Cards[0].Name, specPack.Cards[1].Name)
	}
}

func TestVerifyAllPoolsAgainstGameLua(t *testing.T) {
	data, err := os.ReadFile("../../balatro_src/game.lua")
	if err != nil {
		t.Skip("balatro_src/game.lua not found; skipping Lua pool parity test")
	}
	content := string(data)

	type Center struct {
		Key    string
		Name   string
		Set    string
		Order  int
		Rarity int
	}

	reLine := regexp.MustCompile(`^\s*([a-z0-9_]+)\s*=\s*\{`)
	reName := regexp.MustCompile(`name\s*=\s*(?:"([^"]+)"|'([^']+)')`)
	reSet := regexp.MustCompile(`set\s*=\s*['"]([^'"]+)['"]`)
	reOrder := regexp.MustCompile(`order\s*=\s*([0-9]+)`)
	reRarity := regexp.MustCompile(`rarity\s*=\s*([0-9]+)`)

	var centers []Center
	lines := strings.Split(content, "\n")
	for i := 0; i < len(lines); i++ {
		line := lines[i]
		m := reLine.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		key := m[1]
		// Accumulate lines until closing brace of this entry
		block := line
		for !strings.Contains(block, "},") && !strings.HasSuffix(strings.TrimSpace(block), "}") && i+1 < len(lines) {
			i++
			block += " " + lines[i]
		}

		nameM := reName.FindStringSubmatch(block)
		setM := reSet.FindStringSubmatch(block)
		orderM := reOrder.FindStringSubmatch(block)
		rarityM := reRarity.FindStringSubmatch(block)

		if nameM == nil || setM == nil || orderM == nil {
			continue
		}

		name := nameM[1]
		if name == "" {
			name = nameM[2]
		}

		c := Center{
			Key:   key,
			Name:  name,
			Set:   setM[1],
			Order: 0,
		}
		fmt.Sscanf(orderM[1], "%d", &c.Order)
		if rarityM != nil {
			fmt.Sscanf(rarityM[1], "%d", &c.Rarity)
		}
		centers = append(centers, c)
	}

	t.Logf("Parsed %d centers from game.lua", len(centers))

	// Group by set
	bySet := make(map[string][]Center)
	for _, c := range centers {
		bySet[c.Set] = append(bySet[c.Set], c)
	}

	// 1. Tarots
	tarots := bySet["Tarot"]
	sort.Slice(tarots, func(i, j int) bool { return tarots[i].Order < tarots[j].Order })
	t.Logf("Checking %d Tarots...", len(tarots))
	if len(tarots) != len(TAROTS) {
		t.Errorf("Tarot count mismatch: lua %d vs go %d", len(tarots), len(TAROTS))
	} else {
		for i := range tarots {
			if tarots[i].Name != TAROTS[i] {
				t.Errorf("Tarot [%d] mismatch: lua '%s' vs go '%s'", i, tarots[i].Name, TAROTS[i])
			}
		}
	}

	// 2. Planets
	planets := bySet["Planet"]
	sort.Slice(planets, func(i, j int) bool { return planets[i].Order < planets[j].Order })
	t.Logf("Checking %d Planets...", len(planets))
	if len(planets) != len(PLANETS) {
		t.Errorf("Planet count mismatch: lua %d vs go %d", len(planets), len(PLANETS))
	} else {
		for i := range planets {
			if planets[i].Name != PLANETS[i] {
				t.Errorf("Planet [%d] mismatch: lua '%s' vs go '%s'", i, planets[i].Name, PLANETS[i])
			}
		}
	}

	// 3. Jokers by rarity
	jokers := bySet["Joker"]
	var luaCommon, luaUncommon, luaRare, luaLegendary []string
	sort.Slice(jokers, func(i, j int) bool { return jokers[i].Order < jokers[j].Order })
	for _, j := range jokers {
		switch j.Rarity {
		case 1:
			luaCommon = append(luaCommon, j.Name)
		case 2:
			luaUncommon = append(luaUncommon, j.Name)
		case 3:
			luaRare = append(luaRare, j.Name)
		case 4:
			luaLegendary = append(luaLegendary, j.Name)
		}
	}

	// Compare Commons
	t.Logf("Checking %d Commons...", len(luaCommon))
	if len(luaCommon) != len(COMMON_JOKERS) {
		t.Errorf("Common count mismatch: lua %d vs go %d", len(luaCommon), len(COMMON_JOKERS))
	}
	for i := 0; i < len(luaCommon) && i < len(COMMON_JOKERS); i++ {
		if luaCommon[i] != COMMON_JOKERS[i] {
			t.Errorf("Common [%d] mismatch: lua '%s' vs go '%s'", i, luaCommon[i], COMMON_JOKERS[i])
		}
	}

	// Compare Uncommons
	t.Logf("Checking %d Uncommons...", len(luaUncommon))
	if len(luaUncommon) != len(UNCOMMON_JOKERS) {
		t.Errorf("Uncommon count mismatch: lua %d vs go %d", len(luaUncommon), len(UNCOMMON_JOKERS))
	}
	for i := 0; i < len(luaUncommon) && i < len(UNCOMMON_JOKERS); i++ {
		if luaUncommon[i] != UNCOMMON_JOKERS[i] {
			t.Errorf("Uncommon [%d] mismatch: lua '%s' vs go '%s'", i, luaUncommon[i], UNCOMMON_JOKERS[i])
		}
	}

	// Compare Rares
	t.Logf("Checking %d Rares...", len(luaRare))
	if len(luaRare) != len(RARE_JOKERS) {
		t.Errorf("Rare count mismatch: lua %d vs go %d", len(luaRare), len(RARE_JOKERS))
	}
	for i := 0; i < len(luaRare) && i < len(RARE_JOKERS); i++ {
		if luaRare[i] != RARE_JOKERS[i] {
			t.Errorf("Rare [%d] mismatch: lua '%s' vs go '%s'", i, luaRare[i], RARE_JOKERS[i])
		}
	}

	// Compare Legendaries
	t.Logf("Checking %d Legendaries...", len(luaLegendary))
	if len(luaLegendary) != len(LEGENDARY_JOKERS) {
		t.Errorf("Legendary count mismatch: lua %d vs go %d", len(luaLegendary), len(LEGENDARY_JOKERS))
	}
	for i := 0; i < len(luaLegendary) && i < len(LEGENDARY_JOKERS); i++ {
		if luaLegendary[i] != LEGENDARY_JOKERS[i] {
			t.Errorf("Legendary [%d] mismatch: lua '%s' vs go '%s'", i, luaLegendary[i], LEGENDARY_JOKERS[i])
		}
	}

	// Compare Vouchers
	vouchers := bySet["Voucher"]
	sort.Slice(vouchers, func(i, j int) bool { return vouchers[i].Order < vouchers[j].Order })
	t.Logf("Checking %d Vouchers...", len(vouchers))
	if len(vouchers) != len(VOUCHERS) {
		t.Errorf("Voucher count mismatch: lua %d vs go %d", len(vouchers), len(VOUCHERS))
	} else {
		for i := range vouchers {
			if vouchers[i].Name != VOUCHERS[i] {
				t.Errorf("Voucher [%d] mismatch: lua '%s' vs go '%s'", i, vouchers[i].Name, VOUCHERS[i])
			}
		}
	}
}


