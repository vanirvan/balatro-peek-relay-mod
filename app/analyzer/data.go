package analyzer

// VoucherDef holds base metadata for a voucher in Balatro 1.0.1o.
type VoucherDef struct {
	Key      string
	Name     string
	Order    int
	PosX     int
	PosY     int
	Requires string // empty for tier 1
}

// Master list of all 32 vouchers in Balatro 1.0.1o sorted by order.
var MasterVouchers = []VoucherDef{
	{Key: "v_overstock_norm", Name: "Overstock", Order: 1, PosX: 0, PosY: 0, Requires: ""},
	{Key: "v_overstock_plus", Name: "Overstock Plus", Order: 2, PosX: 0, PosY: 1, Requires: "v_overstock_norm"},
	{Key: "v_clearance_sale", Name: "Clearance Sale", Order: 3, PosX: 3, PosY: 0, Requires: ""},
	{Key: "v_liquidation", Name: "Liquidation", Order: 4, PosX: 3, PosY: 1, Requires: "v_clearance_sale"},
	{Key: "v_hone", Name: "Hone", Order: 5, PosX: 4, PosY: 0, Requires: ""},
	{Key: "v_glow_up", Name: "Glow Up", Order: 6, PosX: 4, PosY: 1, Requires: "v_hone"},
	{Key: "v_reroll_surplus", Name: "Reroll Surplus", Order: 7, PosX: 0, PosY: 2, Requires: ""},
	{Key: "v_reroll_glut", Name: "Reroll Glut", Order: 8, PosX: 0, PosY: 3, Requires: "v_reroll_surplus"},
	{Key: "v_crystal_ball", Name: "Crystal Ball", Order: 9, PosX: 2, PosY: 2, Requires: ""},
	{Key: "v_omen_globe", Name: "Omen Globe", Order: 10, PosX: 2, PosY: 3, Requires: "v_crystal_ball"},
	{Key: "v_telescope", Name: "Telescope", Order: 11, PosX: 3, PosY: 2, Requires: ""},
	{Key: "v_observatory", Name: "Observatory", Order: 12, PosX: 3, PosY: 3, Requires: "v_telescope"},
	{Key: "v_grabber", Name: "Grabber", Order: 13, PosX: 5, PosY: 0, Requires: ""},
	{Key: "v_nacho_tong", Name: "Nacho Tong", Order: 14, PosX: 5, PosY: 1, Requires: "v_grabber"},
	{Key: "v_wasteful", Name: "Wasteful", Order: 15, PosX: 6, PosY: 0, Requires: ""},
	{Key: "v_recyclomancy", Name: "Recyclomancy", Order: 16, PosX: 6, PosY: 1, Requires: "v_wasteful"},
	{Key: "v_tarot_merchant", Name: "Tarot Merchant", Order: 17, PosX: 1, PosY: 0, Requires: ""},
	{Key: "v_tarot_tycoon", Name: "Tarot Tycoon", Order: 18, PosX: 1, PosY: 1, Requires: "v_tarot_merchant"},
	{Key: "v_planet_merchant", Name: "Planet Merchant", Order: 19, PosX: 2, PosY: 0, Requires: ""},
	{Key: "v_planet_tycoon", Name: "Planet Tycoon", Order: 20, PosX: 2, PosY: 1, Requires: "v_planet_merchant"},
	{Key: "v_seed_money", Name: "Seed Money", Order: 21, PosX: 1, PosY: 2, Requires: ""},
	{Key: "v_money_tree", Name: "Money Tree", Order: 22, PosX: 1, PosY: 3, Requires: "v_seed_money"},
	{Key: "v_blank", Name: "Blank", Order: 23, PosX: 7, PosY: 0, Requires: ""},
	{Key: "v_antimatter", Name: "Antimatter", Order: 24, PosX: 7, PosY: 1, Requires: "v_blank"},
	{Key: "v_magic_trick", Name: "Magic Trick", Order: 25, PosX: 4, PosY: 2, Requires: ""},
	{Key: "v_illusion", Name: "Illusion", Order: 26, PosX: 4, PosY: 3, Requires: "v_magic_trick"},
	{Key: "v_hieroglyph", Name: "Hieroglyph", Order: 27, PosX: 5, PosY: 2, Requires: ""},
	{Key: "v_petroglyph", Name: "Petroglyph", Order: 28, PosX: 5, PosY: 3, Requires: "v_hieroglyph"},
	{Key: "v_directors_cut", Name: "Director's Cut", Order: 29, PosX: 6, PosY: 2, Requires: ""},
	{Key: "v_retcon", Name: "Retcon", Order: 30, PosX: 6, PosY: 3, Requires: "v_directors_cut"},
	{Key: "v_paint_brush", Name: "Paint Brush", Order: 31, PosX: 7, PosY: 2, Requires: ""},
	{Key: "v_palette", Name: "Palette", Order: 32, PosX: 7, PosY: 3, Requires: "v_paint_brush"},
}

// MasterBlinds holds the boss blinds eligible in Balatro 1.0.1o.
type BlindDef struct {
	Key      string
	Name     string
	MinAnte  int
	Showdown bool
}

var MasterBossBlinds = []BlindDef{
	{Key: "bl_hook", Name: "The Hook", MinAnte: 1, Showdown: false},
	{Key: "bl_ox", Name: "The Ox", MinAnte: 1, Showdown: false},
	{Key: "bl_house", Name: "The House", MinAnte: 2, Showdown: false},
	{Key: "bl_wheel", Name: "The Wheel", MinAnte: 1, Showdown: false},
	{Key: "bl_arm", Name: "The Arm", MinAnte: 2, Showdown: false},
	{Key: "bl_club", Name: "The Club", MinAnte: 1, Showdown: false},
	{Key: "bl_fish", Name: "The Fish", MinAnte: 2, Showdown: false},
	{Key: "bl_psychic", Name: "The Psychic", MinAnte: 1, Showdown: false},
	{Key: "bl_goad", Name: "The Goad", MinAnte: 1, Showdown: false},
	{Key: "bl_water", Name: "The Water", MinAnte: 2, Showdown: false},
	{Key: "bl_eye", Name: "The Eye", MinAnte: 2, Showdown: false},
	{Key: "bl_mouth", Name: "The Mouth", MinAnte: 2, Showdown: false},
	{Key: "bl_plant", Name: "The Plant", MinAnte: 2, Showdown: false},
	{Key: "bl_serpent", Name: "The Serpent", MinAnte: 2, Showdown: false},
	{Key: "bl_pillar", Name: "The Pillar", MinAnte: 1, Showdown: false},
	{Key: "bl_needle", Name: "The Needle", MinAnte: 2, Showdown: false},
	{Key: "bl_head", Name: "The Head", MinAnte: 1, Showdown: false},
	{Key: "bl_tooth", Name: "The Tooth", MinAnte: 2, Showdown: false},
	{Key: "bl_flint", Name: "The Flint", MinAnte: 2, Showdown: false},
	{Key: "bl_mark", Name: "The Mark", MinAnte: 2, Showdown: false},
	{Key: "bl_wall", Name: "The Wall", MinAnte: 2, Showdown: false},
	{Key: "bl_manacle", Name: "The Manacle", MinAnte: 1, Showdown: false},
	// Showdowns (Ante 8, 16...)
	{Key: "bl_amber_acorn", Name: "Amber Acorn", MinAnte: 8, Showdown: true},
	{Key: "bl_verdant_leaf", Name: "Verdant Leaf", MinAnte: 8, Showdown: true},
	{Key: "bl_violet_vessel", Name: "Violet Vessel", MinAnte: 8, Showdown: true},
	{Key: "bl_crimson_heart", Name: "Crimson Heart", MinAnte: 8, Showdown: true},
	{Key: "bl_cerulean_bell", Name: "Cerulean Bell", MinAnte: 8, Showdown: true},
}

// MasterTags holds the eligible tags in Balatro 1.0.1o.
var MasterTags = []string{
	"tag_uncommon", "tag_rare", "tag_negative", "tag_foil", "tag_holo", "tag_polychrome",
	"tag_investment", "tag_voucher", "tag_boss", "tag_standard", "tag_charm", "tag_meteor",
	"tag_buffoon", "tag_handy", "tag_garbage", "tag_ethereal", "tag_coupon", "tag_double",
	"tag_juggle", "tag_d_six", "tag_top_up", "tag_speed", "tag_orbital", "tag_economy",
}
