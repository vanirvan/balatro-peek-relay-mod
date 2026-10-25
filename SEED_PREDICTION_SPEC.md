# Balatro 1.0.1o PRNG & Seed Prediction Specification

This document provides a comprehensive, reverse-engineered technical specification of the procedural generation (PRNG) architecture used in **Balatro version 1.0.1o-FULL** (running on Love2D / LuaJIT).

It is intended as the primary reference for building stateful seed analyzers and predictor engines (such as in `balatro-peek`), explaining why legacy web tools (*TheSoul*, *Immolate web*) desync, and detailing how future features (shop cards, booster pack contents, Judgement tarot spoilers, Omen Globe modifications) must be implemented.

---

## 1. Core PRNG Architecture

Balatro’s generation consists of **4 distinct layers**:

```
[8-character Seed String] (e.g. "DTCQMCZI")
            │
            ▼
1. pseudohash(str)               ──> Hashing strings into IEEE 754 float64 [0, 1)
            │
            ▼
2. pseudoseed(stream_key)        ──> Stateful per-stream RNG hashing table
            │
            ▼
3. LuaJIT Tausworthe (TW223)    ──> 4-register 64-bit LFSR PRNG (period 2^223)
            │
            ▼
4. Pools & Resample Loops       ──> Filtered item arrays, picking, & '_resample' keys
```

---

### Layer 1: String Hashing (`pseudohash`)
*(Source: `functions/misc_functions.lua:279`)*

Converts an arbitrary ASCII string into a deterministic floating-point number in $[0, 1)$.

```lua
function pseudohash(str)
  local num = 1
  for i = #str, 1, -1 do
      num = ((1.1239285023 / num) * string.byte(str, i) * math.pi + math.pi * i) % 1
  end
  return num
end
```
* **Traversal**: Backwards from the last character to index 1 (1-indexed).
* **Modulo 1**: `num = term - math.Floor(term)`.
* **Precision**: Standard IEEE 754 64-bit float (`double` / `float64`).

---

### Layer 2: Stream State & Seed Derivation (`pseudoseed`)
*(Source: `functions/misc_functions.lua:298`)*

Every RNG event in Balatro uses a named stream identifier (e.g. `'Voucher1'`, `'boss_1'`, `'cdt1'`, `'pack_generic1'`).

```lua
function pseudoseed(key, predict_seed)
  if not G.GAME.pseudorandom[key] then 
    G.GAME.pseudorandom[key] = pseudohash(key .. (G.GAME.pseudorandom.seed or ''))
  end

  G.GAME.pseudorandom[key] = math.abs(tonumber(string.format("%.13f", (2.134453429141 + G.GAME.pseudorandom[key] * 1.72431234) % 1)))
  return (G.GAME.pseudorandom[key] + (G.GAME.pseudorandom.hashed_seed or 0)) / 2
end
```
* **Statefulness**: `G.GAME.pseudorandom[key]` updates in place each time `pseudoseed(key)` is invoked for the same key.
* **Truncation**: `string.format("%.13f", ...)` clamps float precision to exactly 13 decimal digits before taking absolute value.

---

### Layer 3: LuaJIT Tausworthe PRNG (`TW223`)
*(Source: LuaJIT `src/lib_math.c` and `src/lj_prng.c`)*

Balatro uses LuaJIT’s built-in `math.randomseed(d)` and `math.random()`, which implement a 4-register 64-bit Tausworthe LFSR generator with period $2^{223}$.

#### 1. Seeding (`SetRandomSeed(d)`)
Takes the float `d` from `pseudoseed(key)`:
```c
uint32_t r = 0x11090601;
for (int i = 0; i < 4; i++) {
    uint32_t m = 1u << (r & 255);
    r >>= 8;
    d = d * 3.14159265358979323846 + 2.7182818284590452354;
    uint64_t u = *(uint64_t*)&d; // bitwise cast double to uint64
    if (u < m) u += m;
    rs->u[i] = u;
}
// Warm-up: discard 10 steps
for (int i = 0; i < 10; i++) stepTW223();
```

#### 2. Stepping (`stepTW223()`)
Steps 4 registers with specific bit shifts:
* **Step 0**: `k=63, q=31, s=18` (mask: `0xfffffffffffffffe`)
* **Step 1**: `k=58, q=19, s=28` (mask: `0xffffffffffffffc0`)
* **Step 2**: `k=55, q=24, s=7`  (mask: `0xfffffffffffffe00`)
* **Step 3**: `k=47, q=21, s=8`  (mask: `0xfffffffffffe0000`)

```c
z = rs->u[i];
z = (((z << q) ^ z) >> (k - s)) ^ ((z & mask) << s);
r ^= z;
rs->u[i] = z;
```

#### 3. Random Float & Integer Range
* **`Random()`**: `u = (r & 0x000fffffffffffff) | 0x3ff0000000000000; return Float64frombits(u) - 1.0;` $\in [0, 1)$
* **`RandInt(min, max)`**: `floor(Random() * (max - min + 1)) + min`

---

## 2. Item Pools & The "Resample" Mechanism

### How Pools Work
When Balatro selects an item (Vouchers, Jokers, Tarots, Packs), it does **not** shrink the master array when an item is unavailable. 

Instead:
1. It creates an array of identical size to the master list.
2. If an item is eligible, its key is placed in that slot.
3. If an item is ineligible (already owned, unlocked requirement not met, or already in shop), that slot is marked as **`"UNAVAILABLE"`**.

### The Resample Loop
```lua
local center = pseudorandom_element(_pool, pseudoseed(_pool_key))
local it = 1
while center == 'UNAVAILABLE' do
    it = it + 1
    center = pseudorandom_element(_pool, pseudoseed(_pool_key .. '_resample' .. it))
end
```
* **`resample = 1`**: The first roll landed directly on an eligible slot.
* **`resample = 2`**: Roll 1 hit an `"UNAVAILABLE"` slot. The game discarded it, created stream `_pool_key .. '_resample2'`, and rolled again.
* **`resample = n`**: Discarded $n-1$ times until hitting an available slot.

---

## 3. Why Static Analyzers (*TheSoul*, *Immolate*) Desync

Legacy tools like *TheSoul* make rigid assumptions:
1. **The "Forced Buy" Assumption**: They assume the player buys every offered voucher/joker on the ante it appears. If a player **skips** a voucher, it stays in the pool. In a future ante, the game may naturally roll that voucher again on `resample=1`, whereas *TheSoul* assumes it was bought, flags it `UNAVAILABLE`, and resamples to a different voucher.
2. **Ante Rollbacks (Hieroglyph / Petroglyph)**: Buying Hieroglyph drops Ante by 1 (`ease_ante(-1)`). Static loops cannot step backwards or recalculate re-entering an ante with newly unlocked tier-2 vouchers (Petroglyph).
3. **Patch Discrepancies**: 1.0.1f vs 1.0.1o altered voucher requirement checks and Showman handling.

---

## 4. Voucher Mechanics

1. **Master Pool**: 32 vouchers (16 Tier-1, 16 Tier-2).
2. **Key Format**: `'Voucher' .. ante` (e.g. `Voucher1`, `Voucher2`).
3. **Tier-2 Unlocking**: A Tier-2 voucher is only eligible if its Tier-1 prerequisite is present in `G.GAME.used_vouchers`.
4. **Voucher Tag Skipping**:
   - Uses stream key: **`'Voucher_fromtag'`** (independent of the Ante number!).
   - Consecutive Voucher Tags simply advance the `'Voucher_fromtag'` stream.
   - Does **not** consume or affect the normal Ante voucher stream (`'Voucher' .. ante`).

---

## 5. Upcoming Mechanics Implementation Guide

To extend the analyzer with the planned features, follow these exact Balatro 1.0.1o game formulas:

### A. Shop Item Generation (`create_card_for_shop`)
*(Source: `functions/UI_definitions.lua:742`)*

For each card slot in the shop:
1. **Determine Item Type**:
   - Polled against rates: `joker_rate` (20), `tarot_rate` (4), `planet_rate` (4), `playing_card_rate` (0 unless Magic Trick bought = 4), `spectral_rate` (0 unless Ghost Deck = 2).
   - Roll stream: `pseudoseed('cdt' .. ante)`
2. **Generate the Card**:
   - Call `create_card(type, area, nil, nil, nil, nil, nil, 'sho')`
   - Stream key for Jokers: `'Joker' .. rarity .. 'sho' .. ante`
   - Stream key for Tarots: `'Tarotsho' .. ante`
   - Stream key for Planets: `'Planetsho' .. ante`
3. **Joker Stickers (Orange/Gold Stake)**:
   - Eternal/Perishable: `pseudorandom('etperpoll' .. ante)`
   - Rental: `pseudorandom('ssjr' .. ante)`
4. **Joker Editions**:
   - Polled via `poll_edition('edisho' .. ante)`:
     - Foil: 2%, Holographic: 1.4%, Polychrome: 0.3%, Negative: 0.3%

### B. Booster Pack Generation (`get_pack`)
*(Source: `functions/common_events.lua:1944`)*

1. First shop buffoon rule: If no buffoon pack was generated yet, slot 1 is guaranteed to be `p_buffoon_normal_1` or `p_buffoon_normal_2` via `math.random(1, 2)`.
2. Roll booster kind:
   - Cumulative weight of all booster centers in `P_CENTER_POOLS['Booster']`.
   - Roll stream: `pseudoseed('pack_generic' .. ante)`
3. **Booster Card Contents**:
   - **Buffoon Pack**: Generates $N$ jokers using `create_card('Joker', G.pack_cards, nil, nil, nil, nil, nil, 'buf')`.
   - **Arcana Pack**: Generates $N$ tarots using `create_card('Tarot', G.pack_cards, nil, nil, nil, nil, nil, 'ar1')` (or `ar2` for mega).
   - **Celestial Pack**: Generates $N$ planets using `create_card('Planet', G.pack_cards, nil, nil, nil, nil, nil, 'pl1')`.
   - **Spectral Pack**: Generates $N$ spectrals using `create_card('Spectral', G.pack_cards, nil, nil, nil, nil, nil, 'spe')`.
   - **Standard Pack**: Generates playing cards with front suit/rank, enhancement (`'stdcen' .. ante`), seal (`'stdseal' .. ante`), and edition (`'stded' .. ante`).

### C. Omen Globe Effect on Arcana Packs
*(Source: `functions/common_events.lua` & `card.lua`)*

When **Omen Globe** (`v_omen_globe`) is redeemed:
- Spectral cards are added to the eligible pool for Arcana Booster Packs with a weight of 1 in 5 (20% chance to replace a Tarot with a Spectral card).
- When simulating an Arcana pack with Omen Globe owned:
  ```lua
  local card_type = (used_vouchers["v_omen_globe"] and pseudorandom('omen_globe') > 0.8) and 'Spectral' or 'Tarot'
  ```

### D. The Soul & Black Hole Spoilers
*(Source: `functions/common_events.lua:2088`)*

- Inside Arcana or Spectral packs:
  - There is a `0.3%` chance (`pseudorandom('soul_') > 0.997`) to spawn **The Soul** (`c_soul`).
  - When **The Soul** is used, it spawns a Legendary Joker via:
    `pseudorandom_element(G.P_JOKER_RARITY_POOLS[4], pseudoseed('Joker4'))`
  - There is a `0.3%` chance in Celestial packs to spawn **Black Hole** (`c_black_hole`).

### E. Consumable Spoilers (e.g. Judgement, Wraith)
- **Judgement (`c_judgement`)**:
  - Creates a random Joker with default shop rarity distribution using stream:
    `'Joker' .. rarity .. 'jud'`
- **Wraith (`c_wraith`)**:
  - Creates a guaranteed Rare Joker using stream:
    `'Joker3' .. 'wra'`

---

## 6. Architecture in `balatro-peek`

In `balatro-peek`, the analyzer is structured as follows:

```
balatro-peek/
├── mod/
│   └── main.lua              # Steamodded relay: exports live_deck.json with vouchers, jokers, consumables
└── app/
    ├── analyzer/
    │   ├── rng.go            # Pure Go LuaJIT TW223 PRNG & pseudohash
    │   ├── data.go           # Master definitions: vouchers, bosses, tags, jokers, tarots
    │   └── analyzer.go       # Stateful seed predictor (sub-millisecond CPU execution)
    ├── main.go               # Go backend: watches live_deck.json and emits prediction_updated
    └── src/
        ├── mygo.ts           # Type-safe RPC bridge between Go and frontend
        ├── main.ts           # Reactive UI rendering (draw pile, vouchers, predictions)
        └── style.css         # Authentic Balatro CSS shaders and layout
```
