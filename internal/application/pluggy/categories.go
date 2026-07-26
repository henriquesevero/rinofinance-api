package pluggy

import "strings"

// categoryRule maps a substring of Pluggy's (English) transaction category to
// a Portuguese category name, an icon key the frontend understands, and an
// accent color. Rules are matched in order, most specific first.
type categoryRule struct {
	match string
	name  string
	icon  string
	color string
}

var categoryRules = []categoryRule{
	{"credit card payment", "Cartão de crédito", "card", "#EF4444"},
	{"bank slip", "Boletos", "card", "#64748B"},
	{"pix", "Transferências", "card", "#64748B"},
	{"transfer", "Transferências", "card", "#64748B"},
	{"withdrawal", "Saques", "card", "#64748B"},
	{"atm", "Saques", "card", "#64748B"},
	{"salary", "Salário", "work", "#16A34A"},
	{"retirement", "Aposentadoria", "savings", "#16A34A"},
	{"invest", "Investimentos", "savings", "#16A34A"},
	{"interest", "Rendimentos", "savings", "#16A34A"},
	{"fast food", "Alimentação", "utensils", "#F97316"},
	{"restaurant", "Alimentação", "utensils", "#F97316"},
	{"food", "Alimentação", "utensils", "#F97316"},
	{"coffee", "Cafeteria", "coffee", "#B45309"},
	{"grocer", "Supermercado", "cart", "#22C55E"},
	{"market", "Supermercado", "cart", "#22C55E"},
	{"public transport", "Transporte", "bus", "#3B82F6"},
	{"taxi", "Transporte", "car", "#3B82F6"},
	{"ride-hailing", "Transporte", "car", "#3B82F6"},
	{"ride", "Transporte", "car", "#3B82F6"},
	{"parking", "Estacionamento", "car", "#3B82F6"},
	{"automotive", "Automóvel", "car", "#3B82F6"},
	{"fuel", "Combustível", "fuel", "#EA580C"},
	{"gas station", "Combustível", "fuel", "#EA580C"},
	{"electricity", "Contas de casa", "zap", "#EAB308"},
	{"water", "Contas de casa", "home", "#38BDF8"},
	{"utilit", "Contas de casa", "home", "#EAB308"},
	{"internet", "Internet e telefone", "wifi", "#0EA5E9"},
	{"telecom", "Internet e telefone", "wifi", "#0EA5E9"},
	{"mobile", "Internet e telefone", "phone", "#0EA5E9"},
	{"phone", "Internet e telefone", "phone", "#0EA5E9"},
	{"rent", "Moradia", "home", "#8B5CF6"},
	{"housing", "Moradia", "home", "#8B5CF6"},
	{"pharma", "Saúde", "medical", "#EF4444"},
	{"doctor", "Saúde", "medical", "#EF4444"},
	{"health", "Saúde", "health", "#EF4444"},
	{"fitness", "Academia", "gym", "#EF4444"},
	{"gym", "Academia", "gym", "#EF4444"},
	{"education", "Educação", "education", "#6366F1"},
	{"book", "Educação", "books", "#6366F1"},
	{"stream", "Lazer", "movie", "#A855F7"},
	{"movie", "Lazer", "movie", "#A855F7"},
	{"music", "Lazer", "music", "#A855F7"},
	{"game", "Lazer", "game", "#A855F7"},
	{"entertain", "Lazer", "movie", "#A855F7"},
	{"hotel", "Viagem", "travel", "#06B6D4"},
	{"flight", "Viagem", "travel", "#06B6D4"},
	{"travel", "Viagem", "travel", "#06B6D4"},
	{"cloth", "Compras", "bag", "#EC4899"},
	{"electronic", "Compras", "bag", "#EC4899"},
	{"shop", "Compras", "bag", "#EC4899"},
	{"insurance", "Seguros", "tools", "#64748B"},
	{"tax", "Impostos", "tools", "#64748B"},
	{"fee", "Tarifas", "tools", "#64748B"},
	{"pet", "Pets", "pet", "#F59E0B"},
	{"donation", "Doações", "gift", "#EC4899"},
	{"gift", "Presentes", "gift", "#EC4899"},
	{"income", "Renda", "work", "#16A34A"},
}

// mapCategory translates a Pluggy category into a Portuguese category name,
// icon and color. An empty category yields an empty name (uncategorized); an
// unknown one falls back to the original label with a neutral icon so the
// transaction still gets grouped somewhere sensible.
func mapCategory(pluggyCategory string) (name, icon, color string) {
	lc := strings.ToLower(strings.TrimSpace(pluggyCategory))
	if lc == "" {
		return "", "", ""
	}
	for _, r := range categoryRules {
		if strings.Contains(lc, r.match) {
			return r.name, r.icon, r.color
		}
	}
	return strings.TrimSpace(pluggyCategory), "tools", "#64748B"
}
