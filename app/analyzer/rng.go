package analyzer

import (
	"math"
	"strings"
)

var (
	invPrec     = math.Pow(10.0, 13)
	twoInvPrec  = math.Pow(2.0, 13)
	fiveInvPrec = math.Pow(5.0, 13)
)

// Round13 matches Balatro's LocalThunk floating-point rounding to 13 decimal places.
func Round13(x float64) float64 {
	tentative := math.Floor(x*invPrec) / invPrec
	truncated := math.Mod(x*twoInvPrec, 1.0) * fiveInvPrec
	if tentative != x && tentative != math.Nextafter(x, 1.0) && math.Mod(truncated, 1.0) >= 0.5 {
		return (math.Floor(x*invPrec) + 1.0) / invPrec
	}
	return tentative
}

// Pseudohash replicates Balatro's pseudohash(str) string hashing algorithm.
func Pseudohash(s string) float64 {
	num := 1.0
	const pi1 = 3.141592653589793116
	for i := len(s); i >= 1; i-- {
		charVal := float64(s[i-1])
		idx := float64(i)
		term := (1.1239285023/num)*charVal*pi1 + pi1*idx
		num = math.Mod(term, 1.0)
	}
	if math.IsNaN(num) {
		return 0.0
	}
	return num
}

// LuaRandom implements Love2D's / LuaJIT's TW223 Tausworthe LFSR PRNG generator.
type LuaRandom struct {
	state [4]uint64
}

// NewLuaRandom initializes a LuaRandom instance with a numeric seed.
func NewLuaRandom(seed float64) *LuaRandom {
	rng := &LuaRandom{}
	d := seed
	r := uint32(0x11090601)

	for i := 0; i < 4; i++ {
		m := uint64(1) << (r & 255)
		r >>= 8
		d = d*3.14159265358979323846 + 2.7182818284590452354
		u := math.Float64bits(d)
		if u < m {
			u += m
		}
		rng.state[i] = u
	}

	for i := 0; i < 10; i++ {
		rng.stepTW223()
	}

	return rng
}

// stepTW223 steps the 4 LFSR registers.
func (rng *LuaRandom) stepTW223() uint64 {
	var r uint64 = 0

	// Step 0: mask = MAX_UINT64 << 1 = 0xfffffffffffffffe
	{
		z := rng.state[0]
		z = (((z << 31) ^ z) >> 45) ^ ((z & 0xfffffffffffffffe) << 18)
		r ^= z
		rng.state[0] = z
	}

	// Step 1: mask = MAX_UINT64 << 6 = 0xffffffffffffffc0
	{
		z := rng.state[1]
		z = (((z << 19) ^ z) >> 30) ^ ((z & 0xffffffffffffffc0) << 28)
		r ^= z
		rng.state[1] = z
	}

	// Step 2: mask = MAX_UINT64 << 9 = 0xfffffffffffffe00
	{
		z := rng.state[2]
		z = (((z << 24) ^ z) >> 48) ^ ((z & 0xfffffffffffffe00) << 7)
		r ^= z
		rng.state[2] = z
	}

	// Step 3: mask = MAX_UINT64 << 17 = 0xfffffffffffe0000
	{
		z := rng.state[3]
		z = (((z << 21) ^ z) >> 39) ^ ((z & 0xfffffffffffe0000) << 8)
		r ^= z
		rng.state[3] = z
	}

	return r
}

func (rng *LuaRandom) randdblmem() uint64 {
	return (rng.stepTW223() & 0x000fffffffffffff) | 0x3ff0000000000000
}

// Random returns a float64 in [0.0, 1.0) matching Lua's math.random().
func (rng *LuaRandom) Random() float64 {
	return math.Float64frombits(rng.randdblmem()) - 1.0
}

// RandInt returns an integer in [min, max] inclusive.
func (rng *LuaRandom) RandInt(min, max int) int {
	return int(rng.Random()*float64(max-min+1)) + min
}

// CleanSeed sanitizes a seed string to uppercase without whitespace.
func CleanSeed(seed string) string {
	return strings.ToUpper(strings.TrimSpace(seed))
}
