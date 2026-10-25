package analyzer

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

// BalatroRNG replicates Balatro 1.0.1o PRNG and LuaJIT TW223 Tausworthe LFSR generator.
type BalatroRNG struct {
	Seed        string
	HashedSeed  float64
	StreamState map[string]float64
	u           [4]uint64
}

// NewBalatroRNG initializes an RNG instance for a given seed.
func NewBalatroRNG(seed string) *BalatroRNG {
	cleanSeed := strings.ToUpper(strings.TrimSpace(seed))
	return &BalatroRNG{
		Seed:        cleanSeed,
		HashedSeed:  Pseudohash(cleanSeed),
		StreamState: make(map[string]float64),
	}
}

// Clone creates an exact deep copy of the RNG state.
func (b *BalatroRNG) Clone() *BalatroRNG {
	clone := &BalatroRNG{
		Seed:        b.Seed,
		HashedSeed:  b.HashedSeed,
		StreamState: make(map[string]float64, len(b.StreamState)),
		u:           b.u,
	}
	for k, v := range b.StreamState {
		clone.StreamState[k] = v
	}
	return clone
}

// Pseudohash replicates Balatro's pseudohash(str) float arithmetic.
func Pseudohash(str string) float64 {
	num := 1.0
	pi := math.Pi
	for i := len(str) - 1; i >= 0; i-- {
		code := float64(str[i])
		idx := float64(i + 1)
		term := (1.1239285023/num)*code*pi + pi*idx
		num = term - math.Floor(term)
	}
	return num
}

// Pseudoseed advances stream state and produces seed for math.randomseed().
func (b *BalatroRNG) Pseudoseed(key string) float64 {
	state, exists := b.StreamState[key]
	if !exists {
		state = Pseudohash(key + b.Seed)
	}

	raw := 2.134453429141 + state*1.72431234
	raw = raw - math.Floor(raw)

	formatted := fmt.Sprintf("%.13f", raw)
	parsed, _ := strconv.ParseFloat(formatted, 64)
	state = math.Abs(parsed)
	b.StreamState[key] = state

	return (state + b.HashedSeed) / 2.0
}

// SetRandomSeed replicates LuaJIT's math.randomseed(d) warm-up logic.
func (b *BalatroRNG) SetRandomSeed(d float64) {
	r := uint32(0x11090601)
	curD := d

	for i := 0; i < 4; i++ {
		m := uint64(1) << (r & 255)
		r >>= 8

		curD = curD*3.14159265358979323846 + 2.7182818284590452354
		u64 := math.Float64bits(curD)
		if u64 < m {
			u64 += m
		}
		b.u[i] = u64
	}

	for i := 0; i < 10; i++ {
		b.stepTW223()
	}
}

// stepTW223 steps the 4 LFSR registers.
func (b *BalatroRNG) stepTW223() uint64 {
	var r uint64 = 0

	// Step 0: k=63, q=31, s=18 (mask: 64-k = 1) -> 0xfffffffffffffffe
	{
		z := b.u[0]
		mask := uint64(0xfffffffffffffffe)
		z = (((z << 31) ^ z) >> 45) ^ ((z & mask) << 18)
		r ^= z
		b.u[0] = z
	}

	// Step 1: k=58, q=19, s=28 (mask: 64-k = 6) -> 0xffffffffffffffc0
	{
		z := b.u[1]
		mask := uint64(0xffffffffffffffc0)
		z = (((z << 19) ^ z) >> 30) ^ ((z & mask) << 28)
		r ^= z
		b.u[1] = z
	}

	// Step 2: k=55, q=24, s=7 (mask: 64-k = 9) -> 0xfffffffffffffe00
	{
		z := b.u[2]
		mask := uint64(0xfffffffffffffe00)
		z = (((z << 24) ^ z) >> 48) ^ ((z & mask) << 7)
		r ^= z
		b.u[2] = z
	}

	// Step 3: k=47, q=21, s=8 (mask: 64-k = 17) -> 0xfffffffffffe0000
	{
		z := b.u[3]
		mask := uint64(0xfffffffffffe0000)
		z = (((z << 21) ^ z) >> 39) ^ ((z & mask) << 8)
		r ^= z
		b.u[3] = z
	}

	return r
}

// Random returns float64 in [0, 1) matching LuaJIT math.random().
func (b *BalatroRNG) Random() float64 {
	r := b.stepTW223()
	u := (r & 0x000fffffffffffff) | 0x3ff0000000000000
	return math.Float64frombits(u) - 1.0
}

// RandInt returns integer in [min, max] inclusive.
func (b *BalatroRNG) RandInt(min, max int) int {
	d := b.Random()
	return int(math.Floor(d*float64(max-min+1))) + min
}

// PseudorandomElement picks element from string slice with resample support.
func (b *BalatroRNG) PseudorandomElement(pool []string, streamKey string) (string, int) {
	seed := b.Pseudoseed(streamKey)
	b.SetRandomSeed(seed)
	idx := b.RandInt(1, len(pool)) - 1
	return pool[idx], idx
}
