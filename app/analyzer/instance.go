package analyzer

import (
	"fmt"
	"math"
)

// InstParams holds settings for the simulation run.
type InstParams struct {
	Deck      string
	Stake     int // 1 = White, 2 = Red, 3 = Blue, 4 = Purple, 5 = Orange, 6 = Gold, etc.
	Showman   bool
	Vouchers  []string
	Version   int // 10106 for 1.0.1o
	PoolFlags map[string]bool
}

// Instance simulates a Balatro game run's PRNG state.
type Instance struct {
	Seed               string
	HashedSeed         float64
	Nodes              map[string]float64
	Locked             map[string]bool
	Params             InstParams
	GeneratedFirstPack bool
}

// NewInstance creates and initializes an Instance.
func NewInstance(seed string, deck string, stake int, poolFlags map[string]bool) *Instance {
	clean := CleanSeed(seed)
	if deck == "" {
		deck = "Red Deck"
	}
	if stake <= 0 {
		stake = 1
	}

	pfCopy := make(map[string]bool, len(poolFlags))
	for k, v := range poolFlags {
		pfCopy[k] = v
	}

	inst := &Instance{
		Seed:       clean,
		HashedSeed: Pseudohash(clean),
		Nodes:      make(map[string]float64),
		Locked:     make(map[string]bool),
		Params: InstParams{
			Deck:      deck,
			Stake:     stake,
			Showman:   false,
			Vouchers:  make([]string, 0, 16),
			Version:   10106,
			PoolFlags: pfCopy,
		},
		GeneratedFirstPack: false,
	}

	// Starting deck perks
	if deck == "Magic Deck" {
		inst.ActivateVoucher("Crystal Ball")
	} else if deck == "Nebula Deck" {
		inst.ActivateVoucher("Telescope")
	} else if deck == "Zodiac Deck" {
		inst.ActivateVoucher("Tarot Merchant")
		inst.ActivateVoucher("Planet Merchant")
		inst.ActivateVoucher("Overstock")
	}

	return inst
}

// Clone creates an exact deep copy of the Instance state.
func (inst *Instance) Clone() *Instance {
	clone := &Instance{
		Seed:               inst.Seed,
		HashedSeed:         inst.HashedSeed,
		Nodes:              make(map[string]float64, len(inst.Nodes)),
		Locked:             make(map[string]bool, len(inst.Locked)),
		GeneratedFirstPack: inst.GeneratedFirstPack,
		Params: InstParams{
			Deck:      inst.Params.Deck,
			Stake:     inst.Params.Stake,
			Showman:   inst.Params.Showman,
			Version:   inst.Params.Version,
			Vouchers:  make([]string, len(inst.Params.Vouchers)),
			PoolFlags: make(map[string]bool, len(inst.Params.PoolFlags)),
		},
	}
	for k, v := range inst.Nodes {
		clone.Nodes[k] = v
	}
	for k, v := range inst.Locked {
		clone.Locked[k] = v
	}
	copy(clone.Params.Vouchers, inst.Params.Vouchers)
	for k, v := range inst.Params.PoolFlags {
		clone.Params.PoolFlags[k] = v
	}
	return clone
}

// GetNode advances stream state and produces seed for LuaRandom.
func (inst *Instance) GetNode(id string) float64 {
	val, exists := inst.Nodes[id]
	if !exists {
		val = Pseudohash(id + inst.Seed)
		inst.Nodes[id] = val
	}
	val = Round13(math.Mod(val*1.72431234+2.134453429141, 1.0))
	inst.Nodes[id] = val
	return (val + inst.HashedSeed) / 2.0
}

// Random returns a float64 in [0.0, 1.0) for a stream key.
func (inst *Instance) Random(id string) float64 {
	rng := NewLuaRandom(inst.GetNode(id))
	return rng.Random()
}

// RandInt returns an integer in [min, max] inclusive for a stream key.
func (inst *Instance) RandInt(id string, min, max int) int {
	rng := NewLuaRandom(inst.GetNode(id))
	return rng.RandInt(min, max)
}

// RandChoice picks an item from a slice with Balatro resampling if locked.
func (inst *Instance) RandChoice(id string, items []string) string {
	rng := NewLuaRandom(inst.GetNode(id))
	item := items[rng.RandInt(0, len(items)-1)]

	if (!inst.Params.Showman && inst.IsLocked(item)) || item == "RETRY" {
		resample := 2
		for {
			rng = NewLuaRandom(inst.GetNode(fmt.Sprintf("%s_resample%d", id, resample)))
			item = items[rng.RandInt(0, len(items)-1)]
			resample++
			if (item != "RETRY" && (inst.Params.Showman || !inst.IsLocked(item))) || resample > 1000 {
				return item
			}
		}
	}
	return item
}

// RandWeightedChoice picks an item from a weighted pool.
func (inst *Instance) RandWeightedChoice(id string, items []WeightedItem) string {
	rng := NewLuaRandom(inst.GetNode(id))
	poll := rng.Random() * items[0].Weight
	idx := 1
	weight := 0.0
	for weight < poll && idx < len(items) {
		weight += items[idx].Weight
		idx++
	}
	return items[idx-1].Item
}

// Lock marks an item as locked.
func (inst *Instance) Lock(item string) {
	inst.Locked[item] = true
}

// Unlock removes the lock for an item.
func (inst *Instance) Unlock(item string) {
	delete(inst.Locked, item)
}

// IsLocked checks if an item is locked.
func (inst *Instance) IsLocked(item string) bool {
	return inst.Locked[item]
}

// IsVoucherActive checks if a voucher has been redeemed.
func (inst *Instance) IsVoucherActive(voucher string) bool {
	for _, v := range inst.Params.Vouchers {
		if v == voucher {
			return true
		}
	}
	return false
}

// ActivateVoucher redeems a voucher, locks it, and unlocks its Tier 2 upgrade.
func (inst *Instance) ActivateVoucher(voucher string) {
	inst.Params.Vouchers = append(inst.Params.Vouchers, voucher)
	inst.Lock(voucher)

	// Unlock tier 2 voucher counterpart
	for i := 0; i < len(VOUCHERS); i += 2 {
		if VOUCHERS[i] == voucher && i+1 < len(VOUCHERS) {
			inst.Unlock(VOUCHERS[i+1])
			break
		}
	}
}

// InitLocks sets up standard ante and pool locks.
func (inst *Instance) InitLocks(ante int, freshProfile, freshRun bool) {
	if ante < 2 {
		inst.Lock("The Mouth")
		inst.Lock("The Fish")
		inst.Lock("The Wall")
		inst.Lock("The House")
		inst.Lock("The Mark")
		inst.Lock("The Wheel")
		inst.Lock("The Arm")
		inst.Lock("The Water")
		inst.Lock("The Needle")
		inst.Lock("The Flint")
		inst.Lock("Negative Tag")
		inst.Lock("Standard Tag")
		inst.Lock("Meteor Tag")
		inst.Lock("Buffoon Tag")
		inst.Lock("Handy Tag")
		inst.Lock("Garbage Tag")
		inst.Lock("Ethereal Tag")
		inst.Lock("Top-up Tag")
		inst.Lock("Orbital Tag")
	}
	if ante < 3 {
		inst.Lock("The Tooth")
		inst.Lock("The Eye")
	}
	if ante < 4 {
		inst.Lock("The Plant")
	}
	if ante < 5 {
		inst.Lock("The Serpent")
	}
	if ante < 6 {
		inst.Lock("The Ox")
	}

	if freshRun {
		// Tier 2 vouchers are locked until tier 1 is redeemed
		for i := 1; i < len(VOUCHERS); i += 2 {
			inst.Lock(VOUCHERS[i])
		}

		// Unlock secret items if unlocked via pool flags
		if !inst.Params.PoolFlags["planet_x_unlocked"] {
			inst.Lock("Planet X")
		}
		if !inst.Params.PoolFlags["ceres_unlocked"] {
			inst.Lock("Ceres")
		}
		if !inst.Params.PoolFlags["eris_unlocked"] {
			inst.Lock("Eris")
		}
		if !inst.Params.PoolFlags["glass_joker_unlocked"] {
			inst.Lock("Glass Joker")
		}
		if !inst.Params.PoolFlags["cavendish_unlocked"] && !inst.Params.PoolFlags["gros_michel_extinct"] {
			inst.Lock("Cavendish")
		}
		if !inst.Params.PoolFlags["stone_joker_unlocked"] {
			inst.Lock("Stone Joker")
		}
		if !inst.Params.PoolFlags["steel_joker_unlocked"] {
			inst.Lock("Steel Joker")
		}
		if !inst.Params.PoolFlags["lucky_cat_unlocked"] {
			inst.Lock("Lucky Cat")
		}
		if !inst.Params.PoolFlags["golden_ticket_unlocked"] {
			inst.Lock("Golden Ticket")
		}
	}
}

// InitUnlocks lifts locks when advancing to later antes.
func (inst *Instance) InitUnlocks(ante int, freshProfile bool) {
	if ante == 2 {
		inst.Unlock("The Mouth")
		inst.Unlock("The Fish")
		inst.Unlock("The Wall")
		inst.Unlock("The House")
		inst.Unlock("The Mark")
		inst.Unlock("The Wheel")
		inst.Unlock("The Arm")
		inst.Unlock("The Water")
		inst.Unlock("The Needle")
		inst.Unlock("The Flint")
		if !freshProfile {
			inst.Unlock("Negative Tag")
		}
		inst.Unlock("Standard Tag")
		inst.Unlock("Meteor Tag")
		inst.Unlock("Buffoon Tag")
		inst.Unlock("Handy Tag")
		inst.Unlock("Garbage Tag")
		inst.Unlock("Ethereal Tag")
		inst.Unlock("Top-up Tag")
		inst.Unlock("Orbital Tag")
	}
	if ante == 3 {
		inst.Unlock("The Tooth")
		inst.Unlock("The Eye")
	}
	if ante == 4 {
		inst.Unlock("The Plant")
	}
	if ante == 5 {
		inst.Unlock("The Serpent")
	}
	if ante == 6 {
		inst.Unlock("The Ox")
	}
}
