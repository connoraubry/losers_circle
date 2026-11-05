package webgraph

// teamColor is one team's primary (node fill) and secondary (node
// border) brand color, used by the viewer to color graph nodes.
type teamColor struct {
	Primary   string
	Secondary string
}

// nflTeamColors gives each NFL team's primary/secondary brand colors.
// Teams with no entry fall back to the neutral color client-side (see
// app.js).
var nflTeamColors = map[string]teamColor{
	"ARI": {"#97233F", "#000000"},
	"ATL": {"#A71930", "#000000"},
	"BAL": {"#241773", "#000000"},
	"BUF": {"#00338D", "#C60C30"},
	"CAR": {"#0085CA", "#101820"},
	"CHI": {"#0B162A", "#C83803"},
	"CIN": {"#FB4F14", "#000000"},
	"CLE": {"#311D00", "#FF3C00"},
	"DAL": {"#041E42", "#869397"},
	"DEN": {"#FB4F14", "#002244"},
	"DET": {"#0076B6", "#B0B7BC"},
	"GB":  {"#203731", "#FFB612"},
	"HOU": {"#03202F", "#A71930"},
	"IND": {"#002C5F", "#A2AAAD"},
	"JAX": {"#006778", "#D7A22A"},
	"KC":  {"#E31837", "#FFB81C"},
	"LA":  {"#003594", "#FFA300"},
	"LAC": {"#0080C6", "#FFC20E"},
	"LV":  {"#000000", "#A5ACAF"},
	"MIA": {"#008E97", "#FC4C02"},
	"MIN": {"#4F2683", "#FFC62F"},
	"NE":  {"#002244", "#C60C30"},
	"NO":  {"#D3BC8D", "#101820"},
	"NYG": {"#0B2265", "#A71930"},
	"NYJ": {"#125740", "#000000"},
	"PHI": {"#004C54", "#A5ACAF"},
	"PIT": {"#101820", "#FFB612"},
	"SEA": {"#002244", "#69BE28"},
	"SF":  {"#AA0000", "#B3995D"},
	"TB":  {"#D50A0A", "#34302B"},
	"TEN": {"#0C2340", "#4B92DB"},
	"WAS": {"#5A1414", "#FFB612"},
}

// cfbTeamColors gives each tracked CFB (and, reused, CBB) team's
// primary/secondary brand colors, keyed by the same ESPN abbreviation
// cfbconf/cbbconf's membership tables use (the two sports share
// abbreviations for the same school). Covers every team in the
// SEC/Big Ten/ACC/Big 12 membership tables across all tracked seasons,
// including CBB-only members like Notre Dame (ND); teams realigning
// into a tracked conference later need an entry added here too.
var cfbTeamColors = map[string]teamColor{
	// SEC
	"ALA":  {"#9E1B32", "#FFFFFF"},
	"ARK":  {"#9D2235", "#FFFFFF"},
	"AUB":  {"#E87722", "#0C2340"},
	"FLA":  {"#FA4616", "#0021A5"},
	"UGA":  {"#BA0C2F", "#000000"},
	"UK":   {"#0033A0", "#FFFFFF"},
	"LSU":  {"#461D7C", "#FDD023"},
	"MSST": {"#660000", "#FFFFFF"},
	"MIZ":  {"#000000", "#F1B82D"},
	"MISS": {"#CE1126", "#14213D"},
	"SC":   {"#73000A", "#000000"},
	"TENN": {"#FF8200", "#FFFFFF"},
	"TA&M": {"#500000", "#FFFFFF"},
	"VAN":  {"#000000", "#866D4B"},
	"OU":   {"#841617", "#FDF9D8"},
	"TEX":  {"#BF5700", "#FFFFFF"},

	// Big Ten
	"ILL":  {"#FF5F05", "#13294B"},
	"IU":   {"#990000", "#EEEDEB"},
	"IOWA": {"#000000", "#FFCD00"},
	"MD":   {"#E03A3E", "#FFD520"},
	"MICH": {"#FFCB05", "#00274C"},
	"MSU":  {"#18453B", "#FFFFFF"},
	"MINN": {"#7A0019", "#FFCC33"},
	"NEB":  {"#E41C38", "#FDF2D9"},
	"NU":   {"#4E2A84", "#FFFFFF"},
	"OSU":  {"#BB0000", "#666666"},
	"PSU":  {"#041E42", "#FFFFFF"},
	"PUR":  {"#CEB888", "#000000"},
	"RUTG": {"#CC0033", "#FFFFFF"},
	"WIS":  {"#C5050C", "#FFFFFF"},
	"UCLA": {"#2D68C4", "#F2A900"},
	"USC":  {"#990000", "#FFCC00"},
	"ORE":  {"#154733", "#FEE123"},
	"WASH": {"#4B2E83", "#B7A57A"},

	// ACC
	"BC":   {"#98002E", "#BC9B6A"},
	"CLEM": {"#F56600", "#522D80"},
	"DUKE": {"#012169", "#FFFFFF"},
	"FSU":  {"#782F40", "#CEB888"},
	"GT":   {"#B3A369", "#003057"},
	"LOU":  {"#AD0000", "#000000"},
	"MIA":  {"#F47321", "#005030"},
	"NCSU": {"#CC0000", "#FFFFFF"},
	"UNC":  {"#7BAFD4", "#FFFFFF"},
	"PITT": {"#003594", "#FFB81C"},
	"SYR":  {"#F76900", "#000E54"},
	"UVA":  {"#232D4B", "#F84C1E"},
	"VT":   {"#630031", "#CF4420"},
	"WAKE": {"#9E7E38", "#000000"},
	"CAL":  {"#003262", "#FDB515"},
	"SMU":  {"#CC0035", "#354CA1"},
	"STAN": {"#8C1515", "#FFFFFF"},
	"ND":   {"#0C2340", "#AE9142"}, // full ACC member in basketball only; no CFB membership

	// Big 12
	"BAY":  {"#154734", "#FFB81C"},
	"BYU":  {"#002E5D", "#FFFFFF"},
	"CIN":  {"#E00122", "#000000"},
	"HOU":  {"#C8102E", "#FFFFFF"},
	"ISU":  {"#C8102E", "#F1BE48"},
	"KU":   {"#0051BA", "#E8000D"},
	"KSU":  {"#512888", "#FFFFFF"},
	"OKST": {"#FF7300", "#000000"},
	"TCU":  {"#4D1979", "#FFFFFF"},
	"TTU":  {"#CC0000", "#000000"},
	"UCF":  {"#000000", "#FFC904"},
	"WVU":  {"#002855", "#EAAA00"},
	"ARIZ": {"#AB0520", "#0C234B"},
	"ASU":  {"#8C1D40", "#FFC627"},
	"COLO": {"#000000", "#CFB87C"},
	"UTAH": {"#CC0000", "#000000"},
}
