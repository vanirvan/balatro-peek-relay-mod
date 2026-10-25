import { PeekBridge, events, type Card, type LiveState, type SpecialCard, type VoucherItem, type PredictionResult } from "./mygo";

// Balatro 8BitDeck atlas layout (pos: {x, y})
// Each card cell: 71px x 95px
// Suits Y-row: Hearts = 0, Clubs = 1, Diamonds = 2, Spades = 3
// Rank X-col: 2=0, 3=1, 4=2, 5=3, 6=4, 7=5, 8=6, 9=7, 10=8, Jack=9, Queen=10, King=11, Ace=12
const SUIT_ROWS: Record<string, number> = {
  Hearts: 0,
  Clubs: 1,
  Diamonds: 2,
  Spades: 3,
};

const RANK_COLS: Record<string, number> = {
  "2": 0,
  "3": 1,
  "4": 2,
  "5": 3,
  "6": 4,
  "7": 5,
  "8": 6,
  "9": 7,
  "10": 8,
  Jack: 9,
  Queen: 10,
  King: 11,
  Ace: 12,
};

const CARD_W = 71;
const CARD_H = 95;

// Enhancers atlas (994x950 px at 2x -> 497x475 px at 1x; each tile 71x95 px)
// m_bonus: {x:1, y:1}, m_mult: {x:2, y:1}, m_wild: {x:3, y:1}, m_lucky: {x:4, y:1}
// m_glass: {x:5, y:1}, m_steel: {x:6, y:1}, m_stone: {x:5, y:0}, m_gold: {x:6, y:0}
// Default card base (c_base): {x:1, y:0}
const ENHANCEMENT_POS: Record<string, { x: number; y: number }> = {
  "Bonus Card": { x: 1, y: 1 },
  "Bonus": { x: 1, y: 1 },
  "Mult Card": { x: 2, y: 1 },
  "Mult": { x: 2, y: 1 },
  "Wild Card": { x: 3, y: 1 },
  "Lucky Card": { x: 4, y: 1 },
  "Glass Card": { x: 5, y: 1 },
  "Steel Card": { x: 6, y: 1 },
  "Stone Card": { x: 5, y: 0 },
  "Gold Card": { x: 6, y: 0 },
};

const BASE_CARD_POS = { x: 1, y: 0 };

// Seals in Enhancers atlas (shared_seals in Balatro)
// Gold: {x:2, y:0}, Purple: {x:4, y:4}, Red: {x:5, y:4}, Blue: {x:6, y:4}
const SEAL_POS: Record<string, { x: number; y: number }> = {
  Gold: { x: 2, y: 0 },
  Purple: { x: 4, y: 4 },
  Red: { x: 5, y: 4 },
  Blue: { x: 6, y: 4 },
};

// Stickers atlas (stickers.png in Balatro)
// Eternal: {x: 0, y: 0}, Perishable: {x: 0, y: 2}, Rental: {x: 1, y: 2}
const STICKER_POS = {
  eternal: { x: 0, y: 0 },
  perishable: { x: 0, y: 2 },
  rental: { x: 1, y: 2 },
};

function getEnhancerOffset(pos: { x: number; y: number }): { x: number; y: number } {
  return {
    x: -(pos.x * CARD_W),
    y: -(pos.y * CARD_H),
  };
}

function getCardSpriteOffset(suit: string, rank: string): { x: number; y: number } {
  const col = RANK_COLS[rank] ?? 0;
  const row = SUIT_ROWS[suit] ?? 0;
  return {
    x: -(col * CARD_W),
    y: -(row * CARD_H),
  };
}

function renderCard(card: Card, index?: number): HTMLElement {
  const wrapper = document.createElement("div");
  wrapper.className = "card-wrapper";
  if (index !== undefined && index < 5) {
    wrapper.classList.add("is-next-draw");
  }

  if (card.selected) {
    wrapper.classList.add("is-selected");
  }
  if (card.scoring) {
    wrapper.classList.add("is-scoring");
  }

  // Card container stacking base/enhancer behind front suit/rank sprite
  const cardBox = document.createElement("div");
  cardBox.className = "card-box";

  // Layer 1: Base or Enhancer background
  const baseLayer = document.createElement("div");
  baseLayer.className = "card-layer card-base-layer";
  const enhPos = (card.enhancement && ENHANCEMENT_POS[card.enhancement]) || BASE_CARD_POS;
  const enhOffset = getEnhancerOffset(enhPos);
  baseLayer.style.backgroundPosition = `${enhOffset.x}px ${enhOffset.y}px`;
  cardBox.appendChild(baseLayer);

  // Layer 2: Rank & Suit front (skip if stone card, Balatro renders stone without front)
  if (card.enhancement !== "Stone Card" && card.enhancement !== "Stone") {
    const frontLayer = document.createElement("div");
    frontLayer.className = "card-layer card-front-layer";
    const offset = getCardSpriteOffset(card.suit, card.rank);
    frontLayer.style.backgroundPosition = `${offset.x}px ${offset.y}px`;
    cardBox.appendChild(frontLayer);
  }

  // Layer 3: Seal stamp (Gold, Red, Blue, Purple)
  const sealPos = card.seal ? SEAL_POS[card.seal] : undefined;
  if (sealPos) {
    const sealLayer = document.createElement("div");
    sealLayer.className = "card-layer card-seal-layer";
    const sOffset = getEnhancerOffset(sealPos);
    sealLayer.style.backgroundPosition = `${sOffset.x}px ${sOffset.y}px`;
    cardBox.appendChild(sealLayer);
  }

  // Layer 4: Shader Overlay for Edition (Holo, Foil, Polychrome, Negative)
  if (card.edition) {
    const editionLayer = document.createElement("div");
    editionLayer.className = `card-layer card-edition-layer edition-${card.edition.toLowerCase()}`;
    cardBox.appendChild(editionLayer);
  }

  wrapper.appendChild(cardBox);

  // Badge row
  const badgeRow = document.createElement("div");
  badgeRow.className = "card-badge-row";

  if (card.seal) {
    const seal = document.createElement("span");
    seal.className = `badge badge-seal-${card.seal.toLowerCase()}`;
    seal.textContent = card.seal;
    badgeRow.appendChild(seal);
  }

  if (card.edition) {
    const edition = document.createElement("span");
    edition.className = `badge badge-${card.edition.toLowerCase()}`;
    edition.textContent = card.edition;
    badgeRow.appendChild(edition);
  }

  if (card.enhancement) {
    const enh = document.createElement("span");
    enh.className = "badge badge-enhancement";
    enh.textContent = card.enhancement.replace(" Card", "");
    badgeRow.appendChild(enh);
  }

  if (card.debuffed) {
    const debuff = document.createElement("span");
    debuff.className = "badge badge-debuff";
    debuff.textContent = "DEBUFF";
    badgeRow.appendChild(debuff);
  }

  if (badgeRow.children.length > 0) {
    wrapper.appendChild(badgeRow);
  }

  return wrapper;
}

function renderSpecialCard(card: SpecialCard): HTMLElement {
  const wrapper = document.createElement("div");
  wrapper.className = "card-wrapper";

  const cardBox = document.createElement("div");
  cardBox.className = "card-box";

  // Base card background (c_base)
  const baseLayer = document.createElement("div");
  baseLayer.className = "card-layer card-base-layer";
  const baseOffset = getEnhancerOffset(BASE_CARD_POS);
  baseLayer.style.backgroundPosition = `${baseOffset.x}px ${baseOffset.y}px`;
  cardBox.appendChild(baseLayer);

  // Artwork layer: Jokers or Tarots/Planet/Spectral
  const artLayer = document.createElement("div");
  artLayer.className = "card-layer";
  if (card.set === "Joker") {
    artLayer.classList.add("card-joker-layer");
  } else if (card.set === "Voucher") {
    artLayer.classList.add("card-voucher-layer");
  } else {
    artLayer.classList.add("card-tarot-layer");
  }
  artLayer.style.backgroundPosition = `${-(card.pos_x * CARD_W)}px ${-(card.pos_y * CARD_H)}px`;
  cardBox.appendChild(artLayer);

  // Edition shader
  if (card.edition) {
    const editionLayer = document.createElement("div");
    editionLayer.className = `card-layer card-edition-layer edition-${card.edition.toLowerCase()}`;
    cardBox.appendChild(editionLayer);
  }

  // Sticker layers (Eternal, Perishable, Rental)
  if (card.eternal) {
    const stickerLayer = document.createElement("div");
    stickerLayer.className = "card-layer card-sticker-layer";
    stickerLayer.style.backgroundPosition = `${-(STICKER_POS.eternal.x * CARD_W)}px ${-(STICKER_POS.eternal.y * CARD_H)}px`;
    cardBox.appendChild(stickerLayer);
  }
  if (card.perishable) {
    const stickerLayer = document.createElement("div");
    stickerLayer.className = "card-layer card-sticker-layer";
    stickerLayer.style.backgroundPosition = `${-(STICKER_POS.perishable.x * CARD_W)}px ${-(STICKER_POS.perishable.y * CARD_H)}px`;
    cardBox.appendChild(stickerLayer);
  }
  if (card.rental) {
    const stickerLayer = document.createElement("div");
    stickerLayer.className = "card-layer card-sticker-layer";
    stickerLayer.style.backgroundPosition = `${-(STICKER_POS.rental.x * CARD_W)}px ${-(STICKER_POS.rental.y * CARD_H)}px`;
    cardBox.appendChild(stickerLayer);
  }

  // Dynamic Value Pill (e.g. +1 Mult, 1.5x, +50 Chips)
  if (card.val_x_mult !== undefined) {
    const valBadge = document.createElement("span");
    valBadge.className = "joker-val-badge val-badge-xmult";
    valBadge.textContent = `${card.val_x_mult}X`;
    cardBox.appendChild(valBadge);
  } else if (card.val_mult !== undefined) {
    const valBadge = document.createElement("span");
    valBadge.className = "joker-val-badge val-badge-mult";
    valBadge.textContent = `+${card.val_mult}`;
    cardBox.appendChild(valBadge);
  } else if (card.val_chips !== undefined) {
    const valBadge = document.createElement("span");
    valBadge.className = "joker-val-badge val-badge-chips";
    valBadge.textContent = `+${card.val_chips}`;
    cardBox.appendChild(valBadge);
  }

  wrapper.appendChild(cardBox);

  // Label name under card
  if (card.name) {
    const nameLabel = document.createElement("div");
    nameLabel.className = "special-name";
    nameLabel.textContent = card.name;
    nameLabel.title = card.name;
    wrapper.appendChild(nameLabel);
  }

  // Edition badge
  if (card.edition) {
    const badgeRow = document.createElement("div");
    badgeRow.className = "card-badge-row";
    const editionBadge = document.createElement("span");
    editionBadge.className = `badge badge-${card.edition.toLowerCase()}`;
    editionBadge.textContent = card.edition;
    badgeRow.appendChild(editionBadge);
    wrapper.appendChild(badgeRow);
  }

  return wrapper;
}

function updateJokers(jokers: SpecialCard[]) {
  const container = document.getElementById("jokers-list");
  const countBadge = document.getElementById("jokers-count");
  if (!container) return;

  if (countBadge) countBadge.textContent = `${jokers.length} cards`;

  if (jokers.length === 0) {
    container.innerHTML = `<div class="empty-state">No Jokers</div>`;
    return;
  }

  container.innerHTML = "";
  const frag = document.createDocumentFragment();
  jokers.forEach((j) => {
    frag.appendChild(renderSpecialCard(j));
  });
  container.appendChild(frag);
}

function updateConsumables(consumables: SpecialCard[]) {
  const container = document.getElementById("consumables-list");
  const countBadge = document.getElementById("consumables-count");
  if (!container) return;

  if (countBadge) countBadge.textContent = `${consumables.length} cards`;

  if (consumables.length === 0) {
    container.innerHTML = `<div class="empty-state">No Consumables</div>`;
    return;
  }

  container.innerHTML = "";
  const frag = document.createDocumentFragment();
  consumables.forEach((c) => {
    frag.appendChild(renderSpecialCard(c));
  });
  container.appendChild(frag);
}

function updateVouchers(vouchers: VoucherItem[]) {
  const container = document.getElementById("vouchers-list");
  const countBadge = document.getElementById("vouchers-count");
  if (!container) return;

  if (countBadge) countBadge.textContent = `${vouchers.length} vouchers`;

  if (vouchers.length === 0) {
    container.innerHTML = `<div class="empty-state">No vouchers redeemed</div>`;
    return;
  }

  container.innerHTML = "";
  const frag = document.createDocumentFragment();
  vouchers.forEach((v) => {
    const cardLike: SpecialCard = {
      id: v.key,
      name: v.name,
      set: "Voucher",
      pos_x: v.pos_x,
      pos_y: v.pos_y,
      debuffed: false,
      rarity: 1,
      eternal: false,
      perishable: false,
      rental: false,
      pinned: false,
    };
    frag.appendChild(renderSpecialCard(cardLike));
  });
  container.appendChild(frag);
}

function updateHUD(state: LiveState) {
  const seedEl = document.getElementById("val-seed");
  const anteEl = document.getElementById("val-ante");
  const blindEl = document.getElementById("val-blind");
  const handsEl = document.getElementById("val-hands");
  const discardsEl = document.getElementById("val-discards");
  const deckSizeEl = document.getElementById("val-deck-size");

  if (seedEl) seedEl.textContent = state.seed || "--------";
  if (anteEl) anteEl.textContent = String(state.ante);
  if (blindEl) {
    blindEl.textContent = state.blind ? `${state.blind} (${state.blind_chips.toLocaleString()})` : "None";
  }
  if (handsEl) handsEl.textContent = String(state.hands_left);
  if (discardsEl) discardsEl.textContent = String(state.discards_left);
  if (deckSizeEl) deckSizeEl.textContent = String(state.deck_size);
}

function updateDrawPile(cards: Card[]) {
  const container = document.getElementById("draw-pile-list");
  const countBadge = document.getElementById("pile-count");
  if (!container) return;

  if (countBadge) countBadge.textContent = `${cards.length} cards`;

  if (cards.length === 0) {
    container.innerHTML = `<div class="empty-state">Draw pile is empty</div>`;
    return;
  }

  container.innerHTML = "";
  const frag = document.createDocumentFragment();
  cards.forEach((card, idx) => {
    frag.appendChild(renderCard(card, idx));
  });
  container.appendChild(frag);
}

function updateHand(cards: Card[]) {
  const container = document.getElementById("hand-list");
  const countBadge = document.getElementById("hand-count");
  if (!container) return;

  if (countBadge) countBadge.textContent = `${cards.length} cards`;

  if (cards.length === 0) {
    container.innerHTML = `<div class="empty-state">No cards in hand</div>`;
    return;
  }

  container.innerHTML = "";
  const frag = document.createDocumentFragment();
  cards.forEach((card) => {
    frag.appendChild(renderCard(card));
  });
  container.appendChild(frag);
}

function updateScoringBanner(scoringHand?: LiveState["scoring_hand"]) {
  const banner = document.getElementById("scoring-hand-banner");
  const nameEl = document.getElementById("scoring-hand-name");
  const lvlEl = document.getElementById("scoring-hand-lvl");
  const chipsEl = document.getElementById("scoring-chips");
  const multEl = document.getElementById("scoring-mult");

  if (!banner || !nameEl || !lvlEl || !chipsEl || !multEl) return;

  if (!scoringHand || !scoringHand.name) {
    banner.classList.add("hidden");
    return;
  }

  nameEl.textContent = scoringHand.name;
  lvlEl.textContent = `lvl.${scoringHand.level || 1}`;
  chipsEl.textContent = String(scoringHand.chips || 0);
  multEl.textContent = String(scoringHand.mult || 0);
  banner.classList.remove("hidden");
}

function updatePredictions(pred: PredictionResult) {
  const statusEl = document.getElementById("analyzer-status");
  const tagListEl = document.getElementById("tag-queue-items");
  const gridEl = document.getElementById("analyzer-grid");

  if (statusEl) {
    statusEl.textContent = `Ante ${pred.current_ante} Analyzed`;
  }

  // Update Voucher Tag queue preview
  if (tagListEl) {
    if (!pred.next_voucher_tags || pred.next_voucher_tags.length === 0) {
      tagListEl.innerHTML = `<span class="tag-item">None</span>`;
    } else {
      tagListEl.innerHTML = pred.next_voucher_tags
        .map((tag, i) => `<span class="tag-item">#${i + 1}: ${tag}</span>`)
        .join("");
    }
  }

  // Update Ante predictions grid
  if (gridEl) {
    if (!pred.antes || pred.antes.length === 0) {
      gridEl.innerHTML = `<div class="empty-state">No predictions available</div>`;
      return;
    }

    gridEl.innerHTML = "";
    const frag = document.createDocumentFragment();

    pred.antes.forEach((ap) => {
      const card = document.createElement("div");
      card.className = "ante-pred-card";
      if (ap.ante === pred.current_ante) {
        card.classList.add("is-current");
      } else if (ap.ante < pred.current_ante) {
        card.classList.add("is-past");
      }

      // Header: Ante number + Boss
      const header = document.createElement("div");
      header.className = "ante-pred-header";
      header.innerHTML = `
        <span class="ante-num">ANTE ${ap.ante}</span>
        <span class="ante-boss" title="${ap.boss}">${ap.boss}</span>
      `;
      card.appendChild(header);

      // Voucher card widget
      const v = ap.voucher;
      const voucherCard: SpecialCard = {
        id: v.key,
        name: v.name,
        set: "Voucher",
        pos_x: v.pos_x,
        pos_y: v.pos_y,
        debuffed: false,
        rarity: 1,
        eternal: false,
        perishable: false,
        rental: false,
        pinned: false,
      };
      const renderedVoucher = renderSpecialCard(voucherCard);
      card.appendChild(renderedVoucher);

      // Resample info if resampled
      if (v.resample > 1) {
        const resampleEl = document.createElement("span");
        resampleEl.className = "resample-indicator";
        resampleEl.textContent = `(resample #${v.resample})`;
        card.appendChild(resampleEl);
      }

      frag.appendChild(card);
    });

    gridEl.appendChild(frag);
  }
}

function applyState(state: LiveState) {
  updateHUD(state);
  updateScoringBanner(state.scoring_hand);
  updateJokers(state.jokers || []);
  updateConsumables(state.consumables || []);
  updateVouchers(state.vouchers || []);
  updateDrawPile(state.draw_pile || []);
  updateHand(state.hand || []);
}

function initCollapsibleSection(toggleId: string, sectionId: string, storageKey: string, defaultLabel: string) {
  const toggleBtn = document.getElementById(toggleId);
  const section = document.getElementById(sectionId);
  const label = toggleBtn?.querySelector(".collapse-label");
  if (!toggleBtn || !section) return;

  const setCollapsed = (collapsed: boolean) => {
    section.classList.toggle("is-collapsed", collapsed);
    toggleBtn.setAttribute("aria-expanded", String(!collapsed));
    toggleBtn.title = collapsed ? `Expand ${defaultLabel}` : `Collapse ${defaultLabel}`;
    if (label) {
      label.textContent = collapsed ? "Expand" : "Collapse";
    }
    try {
      localStorage.setItem(storageKey, String(collapsed));
    } catch {
      // Ignore storage errors in private browsing
    }
  };

  let initialCollapsed = false;
  try {
    initialCollapsed = localStorage.getItem(storageKey) === "true";
  } catch {
    initialCollapsed = false;
  }

  if (initialCollapsed) {
    setCollapsed(true);
  }

  toggleBtn.addEventListener("click", () => {
    const isCurrentlyCollapsed = section.classList.contains("is-collapsed");
    setCollapsed(!isCurrentlyCollapsed);
  });
}

initCollapsibleSection("draw-pile-toggle", "draw-pile-section", "peek_draw_pile_collapsed", "Draw Pile");
initCollapsibleSection("vouchers-toggle", "vouchers-section", "peek_vouchers_collapsed", "Vouchers");
initCollapsibleSection("analyzer-toggle", "analyzer-section", "peek_analyzer_collapsed", "Seed Predictions");

// Initial pull on window load
PeekBridge.getCurrentState().then((state) => {
  if (state) {
    applyState(state);
    if (state.seed) {
      const vKeys = (state.vouchers || []).map((v) => v.key);
      PeekBridge.getPredictions(state.seed, state.ante, vKeys).then((pred) => {
        if (pred) updatePredictions(pred);
      });
    }
  }
});

// Real-time updates from Go watcher
events.stateUpdated.on((state) => {
  applyState(state);
});

// Real-time predictions pushed from Go watcher
events.predictionUpdated.on((pred) => {
  if (pred) updatePredictions(pred);
});
