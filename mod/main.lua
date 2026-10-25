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

local function extract_card(card, is_selected, is_scoring)
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
        debuffed = card.debuff and true or false,
        selected = is_selected and true or false,
        scoring = is_scoring and true or false
    }
end

local function extract_special_card(card)
    if not card then return nil end

    local center = card.config and card.config.center
    local name = (center and center.name) or (card.ability and card.ability.name) or ""
    local set_type = (center and center.set) or (card.ability and card.ability.set) or ""
    local pos = center and center.pos or {x = 0, y = 0}

    local edition = nil
    if card.edition then
        if card.edition.foil then edition = "foil"
        elseif card.edition.holo then edition = "holo"
        elseif card.edition.polychrome then edition = "polychrome"
        elseif card.edition.negative then edition = "negative"
        elseif type(card.edition.type) == "string" then edition = card.edition.type
        end
    end

    -- Read accurate visual X coordinate (VT.x is rendered transform, T.x is target transform)
    local visual_x = 0
    if card.VT and type(card.VT.x) == "number" then
        visual_x = card.VT.x
    elseif card.T and type(card.T.x) == "number" then
        visual_x = card.T.x
    end

    -- Dynamic scaling values (e.g., mult on Ride the Bus, chips on Ice Cream, x_mult on Constellation)
    local val_mult = nil
    local val_x_mult = nil
    local val_chips = nil
    if card.ability then
        if type(card.ability.mult) == "number" and card.ability.mult > 0 then
            val_mult = card.ability.mult
        end
        if type(card.ability.x_mult) == "number" and card.ability.x_mult > 1 then
            val_x_mult = card.ability.x_mult
        end
        if type(card.ability.chips) == "number" and card.ability.chips > 0 then
            val_chips = card.ability.chips
        end
        -- Some jokers store dynamic mult/chips inside extra
        if type(card.ability.extra) == "table" then
            if type(card.ability.extra.mult) == "number" and card.ability.extra.mult > 0 and not val_mult then
                val_mult = card.ability.extra.mult
            end
            if type(card.ability.extra.chips) == "number" and card.ability.extra.chips > 0 and not val_chips then
                val_chips = card.ability.extra.chips
            end
            if type(card.ability.extra.Xmult) == "number" and card.ability.extra.Xmult > 1 and not val_x_mult then
                val_x_mult = card.ability.extra.Xmult
            end
        end
    end

    local is_eternal = (card.ability and card.ability.eternal) and true or false
    local is_perishable = (card.ability and card.ability.perishable) and true or false
    local is_rental = (card.ability and card.ability.rental) and true or false
    local is_pinned = card.pinned and true or false

    return {
        id = card.unique_val or card.ID or 0,
        name = name,
        set = set_type,
        pos_x = pos.x or 0,
        pos_y = pos.y or 0,
        edition = edition,
        debuffed = card.debuff and true or false,
        rarity = center and center.rarity or 1,
        sort_x = visual_x,
        eternal = is_eternal,
        perishable = is_perishable,
        rental = is_rental,
        pinned = is_pinned,
        val_mult = val_mult,
        val_x_mult = val_x_mult,
        val_chips = val_chips
    }
end

local function dump_live_state()
    if not G or not G.deck or not G.deck.cards then return end

    local draw_pile = {}
    -- G.deck.cards: tail is next drawn card. Iterate backwards so index 1 = next draw.
    local total_cards = #G.deck.cards
    for i = total_cards, 1, -1 do
        local c = extract_card(G.deck.cards[i], false, false)
        if c then
            draw_pile[#draw_pile + 1] = c
        end
    end

    -- Check highlighted/selected cards in hand
    local highlighted_set = {}
    local highlighted_cards = {}
    if G.hand and G.hand.highlighted then
        for i = 1, #G.hand.highlighted do
            local card = G.hand.highlighted[i]
            highlighted_set[card] = true
            highlighted_cards[#highlighted_cards + 1] = card
        end
    end

    -- Evaluate poker hand using official Balatro engine function
    local hand_name = ""
    local hand_level = 0
    local hand_chips = 0
    local hand_mult = 0
    local scoring_set = {}

    if #highlighted_cards > 0 and G.FUNCS and G.FUNCS.get_poker_hand_info then
        local text, loc_disp_text, poker_hands, scoring_hand, disp_text = G.FUNCS.get_poker_hand_info(highlighted_cards)
        if text and text ~= "NULL" then
            hand_name = disp_text or text
            if scoring_hand then
                for _, sc in ipairs(scoring_hand) do
                    scoring_set[sc] = true
                end
            end
            if G.GAME and G.GAME.hands and G.GAME.hands[text] then
                local h_info = G.GAME.hands[text]
                hand_level = h_info.level or 1
                hand_chips = h_info.chips or 0
                hand_mult = h_info.mult or 0
            end
        end
    end

    local hand = {}
    if G.hand and G.hand.cards then
        for i = 1, #G.hand.cards do
            local card = G.hand.cards[i]
            local is_selected = highlighted_set[card] or false
            local is_scoring = scoring_set[card] or false
            local c = extract_card(card, is_selected, is_scoring)
            if c then
                hand[#hand + 1] = c
            end
        end
    end

    -- Extract active Jokers (sorted by visual X position: left to right)
    local jokers = {}
    if G.jokers and G.jokers.cards then
        if G.jokers.sort_cards then pcall(function() G.jokers:sort_cards() end) end
        for i = 1, #G.jokers.cards do
            local jc = extract_special_card(G.jokers.cards[i])
            if jc then
                jokers[#jokers + 1] = jc
            end
        end
        table.sort(jokers, function(a, b) return a.sort_x < b.sort_x end)
    end

    -- Extract active Consumables (Tarots, Planets, Spectrals sorted left to right)
    local consumables = {}
    if G.consumeables and G.consumeables.cards then
        if G.consumeables.sort_cards then pcall(function() G.consumeables:sort_cards() end) end
        for i = 1, #G.consumeables.cards do
            local cc = extract_special_card(G.consumeables.cards[i])
            if cc then
                consumables[#consumables + 1] = cc
            end
        end
        table.sort(consumables, function(a, b) return a.sort_x < b.sort_x end)
    end

    -- Extract redeemed/purchased Vouchers
    local vouchers = {}
    if G.GAME and G.GAME.used_vouchers then
        for v_key, _ in pairs(G.GAME.used_vouchers) do
            local center = G.P_CENTERS and G.P_CENTERS[v_key]
            local name = (center and center.name) or v_key
            local pos = center and center.pos or {x = 0, y = 0}
            local order = center and center.order or 999
            vouchers[#vouchers + 1] = {
                key = v_key,
                name = name,
                set = "Voucher",
                pos_x = pos.x or 0,
                pos_y = pos.y or 0,
                order = order
            }
        end
        table.sort(vouchers, function(a, b) return a.order < b.order end)
    end

    local in_game = G.GAME and G.GAME.round_resets and true or false

    local blind_name = ""
    local blind_chip_target = 0
    if G.GAME and G.GAME.blind then
        blind_name = G.GAME.blind.name or ""
        blind_chip_target = G.GAME.blind.chips or 0
    end

    local payload = {
        timestamp = os.time(),
        status = in_game and "in_game" or "idle",
        seed = (G.GAME and G.GAME.pseudorandom and G.GAME.pseudorandom.seed) or "",
        ante = (G.GAME and G.GAME.round_resets and G.GAME.round_resets.ante) or 0,
        round = (G.GAME and G.GAME.round) or 0,
        blind = blind_name,
        blind_chips = blind_chip_target,
        hands_left = (G.GAME and G.GAME.current_round and G.GAME.current_round.hands_left) or 0,
        discards_left = (G.GAME and G.GAME.current_round and G.GAME.current_round.discards_left) or 0,
        deck_size = total_cards,
        scoring_hand = {
            name = hand_name,
            level = hand_level,
            chips = hand_chips,
            mult = hand_mult
        },
        jokers = jokers,
        consumables = consumables,
        vouchers = vouchers,
        draw_pile = draw_pile,
        hand = hand
    }

    local json_str = to_json(payload)

    -- Quick change-detect to avoid redundant disk writes
    local hash = string.format("%d_%d_%d_%d_%d_%d_%d_%d_%s_%s", total_cards, #hand, #jokers, #consumables, #vouchers, #highlighted_cards, payload.hands_left, payload.discards_left, blind_name, hand_name)
    if draw_pile[1] then
        hash = hash .. "_" .. tostring(draw_pile[1].id)
    end
    for i = 1, #jokers do
        local j = jokers[i]
        hash = hash .. "_j" .. tostring(j.id) .. "_" .. tostring(j.val_mult or 0) .. "_" .. tostring(j.val_chips or 0) .. "_" .. tostring(j.val_x_mult or 0)
    end
    for i = 1, #consumables do
        hash = hash .. "_c" .. tostring(consumables[i].id)
    end
    for i = 1, #vouchers do
        hash = hash .. "_v" .. tostring(vouchers[i].key)
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
        local ok, err = pcall(dump_live_state)
        if not ok then
            print("[PeekRelay ERROR] " .. tostring(err))
        end
    end
end
