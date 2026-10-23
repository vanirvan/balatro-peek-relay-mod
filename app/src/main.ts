import { PeekBridge, events, type Card, type LiveState } from "./mygo";

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

function applyState(state: LiveState) {
  updateHUD(state);
  updateScoringBanner(state.scoring_hand);
  updateDrawPile(state.draw_pile || []);
  updateHand(state.hand || []);
}

// Initial pull on window load
PeekBridge.getCurrentState().then((state) => {
  if (state) applyState(state);
});

// Real-time updates from Go watcher
events.stateUpdated.on((state) => {
  applyState(state);
});
