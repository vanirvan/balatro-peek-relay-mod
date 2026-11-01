import {
  PeekBridge,
  events,
  type AntePrediction,
  type Card,
  type CardItem,
  type LiveState,
  type PredictionResult,
  type SpecialCard,
  type TagPrediction,
  type VoucherItem,
} from "./mygo";

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
  Bonus: { x: 1, y: 1 },
  "Mult Card": { x: 2, y: 1 },
  Mult: { x: 2, y: 1 },
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

// Render predicted card item (in packs, tag rewards, etc.)
function renderCardItem(card: CardItem): HTMLElement {
  const wrapper = document.createElement("div");
  wrapper.className = "card-wrapper";

  const cardBox = document.createElement("div");
  cardBox.className = "card-box";

  // Check card type: Playing card vs Special card (Joker/Tarot/Planet/Spectral/Voucher)
  if (card.set === "Default" || card.set === "Enhanced" || (card.rank && card.suit)) {
    // Base or Enhancer background
    const baseLayer = document.createElement("div");
    baseLayer.className = "card-layer card-base-layer";
    const enhPos = (card.enhancement && ENHANCEMENT_POS[card.enhancement]) || BASE_CARD_POS;
    const enhOffset = getEnhancerOffset(enhPos);
    baseLayer.style.backgroundPosition = `${enhOffset.x}px ${enhOffset.y}px`;
    cardBox.appendChild(baseLayer);

    // Front rank/suit layer (skip if stone)
    if (card.enhancement !== "Stone Card" && card.enhancement !== "Stone" && card.suit && card.rank) {
      const frontLayer = document.createElement("div");
      frontLayer.className = "card-layer card-front-layer";
      const offset = getCardSpriteOffset(card.suit, card.rank);
      frontLayer.style.backgroundPosition = `${offset.x}px ${offset.y}px`;
      cardBox.appendChild(frontLayer);
    }

    // Seal
    const sealPos = card.seal ? SEAL_POS[card.seal] : undefined;
    if (sealPos) {
      const sealLayer = document.createElement("div");
      sealLayer.className = "card-layer card-seal-layer";
      const sOffset = getEnhancerOffset(sealPos);
      sealLayer.style.backgroundPosition = `${sOffset.x}px ${sOffset.y}px`;
      cardBox.appendChild(sealLayer);
    }
  } else {
    // Special card (Joker, Tarot, Planet, Spectral, Voucher)
    const baseLayer = document.createElement("div");
    baseLayer.className = "card-layer card-base-layer";
    const baseOffset = getEnhancerOffset(BASE_CARD_POS);
    baseLayer.style.backgroundPosition = `${baseOffset.x}px ${baseOffset.y}px`;
    cardBox.appendChild(baseLayer);

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
  }

  // Edition overlay
  if (card.edition) {
    const editionLayer = document.createElement("div");
    editionLayer.className = `card-layer card-edition-layer edition-${card.edition.toLowerCase()}`;
    cardBox.appendChild(editionLayer);
  }

  // Stickers (Eternal, Perishable, Rental)
  if (card.stickers?.eternal) {
    const sLayer = document.createElement("div");
    sLayer.className = "card-layer card-sticker-layer";
    sLayer.style.backgroundPosition = `${-(STICKER_POS.eternal.x * CARD_W)}px ${-(STICKER_POS.eternal.y * CARD_H)}px`;
    cardBox.appendChild(sLayer);
  }
  if (card.stickers?.perishable) {
    const sLayer = document.createElement("div");
    sLayer.className = "card-layer card-sticker-layer";
    sLayer.style.backgroundPosition = `${-(STICKER_POS.perishable.x * CARD_W)}px ${-(STICKER_POS.perishable.y * CARD_H)}px`;
    cardBox.appendChild(sLayer);
  }
  if (card.stickers?.rental) {
    const sLayer = document.createElement("div");
    sLayer.className = "card-layer card-sticker-layer";
    sLayer.style.backgroundPosition = `${-(STICKER_POS.rental.x * CARD_W)}px ${-(STICKER_POS.rental.y * CARD_H)}px`;
    cardBox.appendChild(sLayer);
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

  if (card.stickers?.eternal) {
    const st = document.createElement("span");
    st.className = "badge badge-debuff";
    st.textContent = "ETERNAL";
    badgeRow.appendChild(st);
  }
  if (card.stickers?.perishable) {
    const st = document.createElement("span");
    st.className = "badge badge-debuff";
    st.textContent = "PERISH";
    badgeRow.appendChild(st);
  }
  if (card.stickers?.rental) {
    const st = document.createElement("span");
    st.className = "badge badge-debuff";
    st.textContent = "RENTAL";
    badgeRow.appendChild(st);
  }

  if (badgeRow.children.length > 0) {
    wrapper.appendChild(badgeRow);
  }

  return wrapper;
}

function renderTagBadge(tag: TagPrediction): HTMLElement {
  const card = document.createElement("div");
  card.className = "tag-badge-card";

  // Chip sprite (34x34 px)
  const chip = document.createElement("div");
  chip.className = "tag-chip-sprite";
  chip.style.backgroundPosition = `${-(tag.pos_x * 34)}px ${-(tag.pos_y * 34)}px`;
  chip.title = `${tag.name} (${tag.blind} Blind)`;
  card.appendChild(chip);

  // Labels column
  const labelCol = document.createElement("div");
  labelCol.className = "tag-label-col";

  const titleRow = document.createElement("div");
  titleRow.className = "tag-title-row";

  const blindPill = document.createElement("span");
  blindPill.className = "tag-blind-pill";
  blindPill.textContent = tag.blind;
  titleRow.appendChild(blindPill);

  const tagName = document.createElement("span");
  tagName.className = "tag-name";
  tagName.textContent = tag.name;
  titleRow.appendChild(tagName);

  labelCol.appendChild(titleRow);

  const rewardDesc = document.createElement("span");
  rewardDesc.className = "tag-reward-desc";
  rewardDesc.textContent = tag.reward_title;
  labelCol.appendChild(rewardDesc);

  // If there are reward preview cards (e.g. Charm Tag gives 5 Arcana cards, Top-up gives Jokers, etc.)
  if (tag.reward_cards && tag.reward_cards.length > 0) {
    const previewContainer = document.createElement("div");
    previewContainer.className = "tag-reward-preview-cards";
    tag.reward_cards.forEach((rc) => {
      const miniChip = document.createElement("span");
      miniChip.className = "tag-reward-mini-chip";
      let text = rc.name;
      if (rc.edition) text += ` (${rc.edition})`;
      miniChip.textContent = text;
      previewContainer.appendChild(miniChip);
    });
    labelCol.appendChild(previewContainer);
  }

  card.appendChild(labelCol);
  return card;
}

// Global prediction state
let latestLiveState: LiveState | null = null;
let latestPrediction: PredictionResult | null = null;
let selectedAnte: number = 1;
let lastKnownAnte: number = -1;

function renderActiveAnte() {
  const container = document.getElementById("ante-prediction-view");
  const badge = document.getElementById("prediction-status-badge");
  if (!container) return;

  if (!latestPrediction || !latestPrediction.antes || latestPrediction.antes.length === 0) {
    container.innerHTML = `<div class="empty-state">Waiting for seed and game state...</div>`;
    return;
  }

  const anteData = latestPrediction.antes.find((a) => a.ante === selectedAnte) || latestPrediction.antes[0];
  if (!anteData) {
    container.innerHTML = `<div class="empty-state">No prediction for Ante ${selectedAnte}</div>`;
    return;
  }

  if (badge) {
    badge.textContent = `Ante ${anteData.ante}`;
  }

  container.innerHTML = "";

  // Main Ante Title Header
  const titleHeader = document.createElement("div");
  titleHeader.className = "ante-main-title";
  titleHeader.innerHTML = `<span>Ante ${anteData.ante}</span>`;
  container.appendChild(titleHeader);

  // Top Row: Voucher Card on left + Modifiers (Boss Blind & Tags) on right
  const topRow = document.createElement("div");
  topRow.className = "ante-top-row";

  // 1. Large Voucher Card Box on the left
  const voucherCol = document.createElement("div");
  voucherCol.className = "voucher-col";

  const voucherBox = document.createElement("div");
  voucherBox.className = "voucher-card-box";
  voucherBox.title = anteData.voucher.name;

  const vBase = document.createElement("div");
  vBase.className = "card-layer card-base-layer";
  const baseOffset = getEnhancerOffset(BASE_CARD_POS);
  vBase.style.backgroundPosition = `${baseOffset.x}px ${baseOffset.y}px`;
  voucherBox.appendChild(vBase);

  const vArt = document.createElement("div");
  vArt.className = "card-layer card-voucher-layer";
  vArt.style.backgroundPosition = `${-(anteData.voucher.pos_x * CARD_W)}px ${-(anteData.voucher.pos_y * CARD_H)}px`;
  voucherBox.appendChild(vArt);

  voucherCol.appendChild(voucherBox);

  const vName = document.createElement("div");
  vName.className = "special-name";
  vName.textContent = anteData.voucher.name;
  voucherCol.appendChild(vName);

  topRow.appendChild(voucherCol);

  // 2. Modifiers Column (Boss Blind & Tags)
  const modCol = document.createElement("div");
  modCol.className = "ante-modifiers-col";

  // Boss Blind group
  const bossGroup = document.createElement("div");
  bossGroup.className = "modifier-badge-group";

  const bossCard = document.createElement("div");
  bossCard.className = "boss-info-card";

  const bossChip = document.createElement("div");
  bossChip.className = "boss-chip-sprite";
  bossChip.style.backgroundPosition = `${-(anteData.boss.pos_x * 34)}px ${-(anteData.boss.pos_y * 34)}px`;
  bossChip.title = anteData.boss.name;
  bossCard.appendChild(bossChip);

  const bossName = document.createElement("span");
  bossName.className = "boss-name-text";
  bossName.textContent = anteData.boss.name;
  bossCard.appendChild(bossName);

  bossGroup.appendChild(bossCard);

  // Director's Cut 1st Reroll
  if (anteData.boss.director_cut_reroll) {
    const dcPill = document.createElement("div");
    dcPill.className = "boss-reroll-pill";
    dcPill.title = "Director's Cut Voucher 1st Reroll";
    dcPill.innerHTML = `<span>🎲 Reroll 1:</span><strong>${anteData.boss.director_cut_reroll.name}</strong>`;
    bossGroup.appendChild(dcPill);
  }

  // Retcon Voucher 2nd & 3rd Rerolls
  if (anteData.boss.retcon_rerolls && anteData.boss.retcon_rerolls.length > 0) {
    anteData.boss.retcon_rerolls.forEach((r, idx) => {
      const retconPill = document.createElement("div");
      retconPill.className = "boss-reroll-pill retcon";
      retconPill.title = `Retcon Voucher Reroll ${idx + 2}`;
      retconPill.innerHTML = `<span>🎲 Reroll ${idx + 2}:</span><strong>${r.name}</strong>`;
      bossGroup.appendChild(retconPill);
    });
  }

  modCol.appendChild(bossGroup);

  // Tags group
  if (anteData.tags && anteData.tags.length > 0) {
    const tagsGroup = document.createElement("div");
    tagsGroup.className = "modifier-badge-group";
    anteData.tags.forEach((tag) => {
      tagsGroup.appendChild(renderTagBadge(tag));
    });
    modCol.appendChild(tagsGroup);
  }

  topRow.appendChild(modCol);
  container.appendChild(topRow);

  // Packs Section
  const packsSection = document.createElement("div");
  packsSection.className = "packs-section-container";

  const packsHeader = document.createElement("h3");
  packsHeader.className = "packs-main-header";
  packsHeader.textContent = "Packs";
  packsSection.appendChild(packsHeader);

  anteData.packs.forEach((blindPack) => {
    const blindGroup = document.createElement("div");
    blindGroup.className = "blind-packs-group";

    const blindTitle = document.createElement("div");
    blindTitle.className = "blind-group-title";
    blindTitle.textContent = blindPack.blind;
    blindGroup.appendChild(blindTitle);

    if (blindPack.packs.length === 0) {
      const noPacks = document.createElement("div");
      noPacks.className = "no-packs-text";
      noPacks.textContent = "No packs available";
      blindGroup.appendChild(noPacks);
    } else {
      blindPack.packs.forEach((pack) => {
        const packEntry = document.createElement("div");
        packEntry.className = "pack-entry";

        const packTitle = document.createElement("div");
        packTitle.className = "pack-entry-title";
        packTitle.textContent = pack.title;
        packEntry.appendChild(packTitle);

        const cardsRow = document.createElement("div");
        cardsRow.className = "pack-cards-row";
        pack.cards.forEach((c) => {
          cardsRow.appendChild(renderCardItem(c));
        });
        packEntry.appendChild(cardsRow);

        blindGroup.appendChild(packEntry);
      });
    }

    packsSection.appendChild(blindGroup);
  });

  container.appendChild(packsSection);
}

function updateAnteTabs(pred: PredictionResult) {
  const tabsBar = document.getElementById("ante-tabs");
  if (!tabsBar) return;

  tabsBar.innerHTML = "";
  pred.antes.forEach((a) => {
    const tab = document.createElement("button");
    tab.className = "ante-tab";
    tab.type = "button";
    tab.textContent = `Ante ${a.ante}`;
    if (a.ante === selectedAnte) {
      tab.classList.add("is-active");
      requestAnimationFrame(() => {
        tab.scrollIntoView({ behavior: "smooth", block: "nearest", inline: "nearest" });
      });
    }
    if (latestLiveState && latestLiveState.ante === a.ante) {
      tab.classList.add("is-current-run");
    }
    tab.addEventListener("click", () => {
      selectedAnte = a.ante;
      updateAnteTabs(pred);
      renderActiveAnte();
    });
    tabsBar.appendChild(tab);
  });
}

function applyPredictions(pred: PredictionResult) {
  latestPrediction = pred;
  if (latestLiveState && latestLiveState.ante > 0) {
    selectedAnte = latestLiveState.ante;
  } else if (!selectedAnte || selectedAnte < 1 || selectedAnte > pred.antes.length) {
    selectedAnte = pred.current_ante || 1;
  }
  updateAnteTabs(pred);
  renderActiveAnte();
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

let currentSeed: string = "";

function updateHUD(state: LiveState) {
  currentSeed = state.seed || "";
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
  const anteChanged = state.ante > 0 && state.ante !== lastKnownAnte;
  if (anteChanged) {
    lastKnownAnte = state.ante;
    selectedAnte = state.ante;
  }
  latestLiveState = state;
  updateHUD(state);
  updateScoringBanner(state.scoring_hand);
  updateJokers(state.jokers || []);
  updateConsumables(state.consumables || []);
  updateVouchers(state.vouchers || []);
  updateDrawPile(state.draw_pile || []);
  updateHand(state.hand || []);

  if (latestPrediction) {
    if (anteChanged) {
      renderActiveAnte();
    }
    updateAnteTabs(latestPrediction);
  } else if (state.seed) {
    // Initial fetch if prediction not yet received
    const vKeys = (state.vouchers || []).map((v) => v.key);
    const jNames = (state.jokers || []).map((j) => j.name);
    const cNames = (state.consumables || []).map((c) => c.name);
    PeekBridge.getPredictions(
      state.seed,
      state.ante,
      vKeys,
      jNames,
      cNames,
      state.pool_flags || {},
      state.deck || "Red Deck",
      state.stake || 1
    ).then((pred) => {
      if (pred) applyPredictions(pred);
    });
  }
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

function initSeedCopyButton() {
  const btn = document.getElementById("hud-seed-btn");
  const toast = document.getElementById("seed-copied-toast");
  if (!btn) return;

  let toastTimer: number | null = null;

  const copy = async () => {
    const seed = currentSeed || document.getElementById("val-seed")?.textContent?.trim() || "";
    if (!seed || seed === "--------") return;

    let copied = false;
    try {
      if (navigator.clipboard && navigator.clipboard.writeText) {
        await navigator.clipboard.writeText(seed);
        copied = true;
      }
    } catch {
      copied = false;
    }

    if (!copied) {
      try {
        const textarea = document.createElement("textarea");
        textarea.value = seed;
        textarea.style.position = "fixed";
        textarea.style.opacity = "0";
        document.body.appendChild(textarea);
        textarea.focus();
        textarea.select();
        copied = document.execCommand("copy");
        document.body.removeChild(textarea);
      } catch {
        copied = false;
      }
    }

    if (copied && toast) {
      toast.classList.add("visible");
      if (toastTimer !== null) clearTimeout(toastTimer);
      toastTimer = window.setTimeout(() => {
        toast.classList.remove("visible");
        toastTimer = null;
      }, 1200);
    }
  };

  btn.addEventListener("click", copy);
  btn.addEventListener("keydown", (e) => {
    if (e.key === "Enter" || e.key === " ") {
      e.preventDefault();
      copy();
    }
  });
}

initCollapsibleSection("draw-pile-toggle", "draw-pile-section", "peek_draw_pile_collapsed", "Draw Pile");
initCollapsibleSection("vouchers-toggle", "vouchers-section", "peek_vouchers_collapsed", "Vouchers");
initCollapsibleSection("prediction-toggle", "prediction-section", "peek_prediction_collapsed", "Seed Prediction");
initSeedCopyButton();

// Initial pull on window load
PeekBridge.getCurrentState().then((state) => {
  if (state) applyState(state);
});

// Real-time updates from Go watcher
events.stateUpdated.on((state) => {
  applyState(state);
});

// Seed prediction updates from Go analyzer
events.predictionUpdated.on((pred) => {
  applyPredictions(pred);
});
