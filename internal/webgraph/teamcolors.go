package webgraph

// teamColor is one team's primary (node fill) and secondary (node
// border) brand color, used by the viewer to color graph nodes.
type teamColor struct {
	Primary   string
	Secondary string
}

// nflTeamColors gives each NFL team's primary/secondary brand colors.
// CFB teams have no entry yet, so their nodes fall back to the neutral
// color client-side (see app.js).
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
