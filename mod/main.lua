--- STEAMODDED HEADER
--- ModName: Peek Relay
--- ModID: PeekRelay
--- ModVersion: 1.0.0
--- ModAuthor: Irvan
--- ModDescription: Bridge mod: exports real-time draw pile and game state to live_deck.json
--- Prefix: peekrelay

local update_timer = 0
local UPDATE_INTERVAL = 0.25
local last_dump_hash = ""

-- Minimal fast JSON serializer (no external dependency)
local function to_json(val)
    local t = type(val)
    if val == nil then
        return "null"
    elseif t == "boolean" then
        return val and "true" or "false"
    elseif t == "number" then
        if val ~= val or val == math.huge or val == -math.huge then
            return "null"
        end
        return tostring(val)
    elseif t == "string" then
        return string.format("%q", val)
    elseif t == "table" then
        local is_array = true
        local n = 0
        for k, _ in pairs(val) do
            n = n + 1
            if type(k) ~= "number" or k < 1 or math.floor(k) ~= k then
                is_array = false
                break
            end
        end
        if is_array and n > 0 then
            for i = 1, n do
                if val[i] == nil then
                    is_array = false
                    break
                end
            end
        end

        local parts = {}
        if is_array then
            for i = 1, #val do
                parts[#parts + 1] = to_json(val[i])
            end
            return "[" .. table.concat(parts, ",") .. "]"
        else
            for k, v in pairs(val) do
                parts[#parts + 1] = string.format("%q:%s", tostring(k), to_json(v))
            end
            return "{" .. table.concat(parts, ",") .. "}"
        end
    end
    return "null"
end

local function extract_card(card)
    if not card then return nil end

    local rank = (card.base and card.base.value) or nil
    local suit = (card.base and card.base.suit) or nil
    local id = card.unique_val or card.ID or 0

    local seal = card.seal or nil

    local enhancement = nil
    if card.config and card.config.center and card.config.center.key and card.config.center.key ~= "c_base" then
        enhancement = card.config.center.name or card.config.center.key
    end

    local edition = nil
    if card.edition then
        if card.edition.foil then edition = "foil"
        elseif card.edition.holo then edition = "holo"
        elseif card.edition.polychrome then edition = "polychrome"
        elseif card.edition.negative then edition = "negative"
        elseif type(card.edition.type) == "string" then edition = card.edition.type
        end
    end

    return {
        id = id,
        rank = rank,
        suit = suit,
        seal = seal,
        enhancement = enhancement,
        edition = edition,
        debuffed = card.debuff and true or false
    }
end

local function dump_live_state()
    if not G or not G.deck or not G.deck.cards then return end

    local draw_pile = {}
    -- G.deck.cards: tail is next drawn card. Iterate backwards so index 1 = next draw.
    local total_cards = #G.deck.cards
    for i = total_cards, 1, -1 do
        local c = extract_card(G.deck.cards[i])
        if c then
            draw_pile[#draw_pile + 1] = c
        end
    end

    local hand = {}
    if G.hand and G.hand.cards then
        for i = 1, #G.hand.cards do
            local c = extract_card(G.hand.cards[i])
            if c then
                hand[#hand + 1] = c
            end
        end
    end

    local in_game = G.GAME and G.GAME.round_resets and true or false

    local payload = {
        timestamp = os.time(),
        status = in_game and "in_game" or "idle",
        seed = (G.GAME and G.GAME.pseudorandom and G.GAME.pseudorandom.seed) or "",
        ante = (G.GAME and G.GAME.round_resets and G.GAME.round_resets.ante) or 0,
        round = (G.GAME and G.GAME.round) or 0,
        hands_left = (G.GAME and G.GAME.current_round and G.GAME.current_round.hands_left) or 0,
        discards_left = (G.GAME and G.GAME.current_round and G.GAME.current_round.discards_left) or 0,
        deck_size = total_cards,
        draw_pile = draw_pile,
        hand = hand
    }

    local json_str = to_json(payload)

    -- Quick change-detect to avoid redundant disk writes
    local hash = string.format("%d_%d_%d_%d", total_cards, #hand, payload.hands_left, payload.discards_left)
    if draw_pile[1] then
        hash = hash .. "_" .. tostring(draw_pile[1].id)
    end

    if hash ~= last_dump_hash then
        last_dump_hash = hash
        -- love.filesystem writes relative to game save directory (%AppData%/Balatro/)
        love.filesystem.write("Mods/PeekRelay/live_deck.json", json_str)
    end
end

-- Hook Game:update
local game_update_ref = Game.update
function Game:update(dt)
    game_update_ref(self, dt)

    update_timer = update_timer + dt
    if update_timer >= UPDATE_INTERVAL then
        update_timer = 0
        pcall(dump_live_state)
    end
end
