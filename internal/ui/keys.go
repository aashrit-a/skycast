package ui

// Key names as reported by tea.KeyMsg.String().
var (
	keysQuit    = []string{"q", "esc", "ctrl+c"}
	keysRefresh = []string{"r"}
	keysUnits   = []string{"u"}
	keysPanel   = []string{"p"}
	keysAmbient = []string{"a"}
	keysHelp    = []string{"?"}
	keysNext    = []string{"right", "n"}
	keysPrev    = []string{"left", "b"}
)

// footerKeys is the key reference shown in the footer.
var footerKeys = []string{"q quit", "r refresh", "u units", "p panel", "a ambient", "? help"}

// footerDemoKeys replaces refresh with scene navigation in --demo mode.
var footerDemoKeys = []string{"←/→ scene", "q quit", "u units", "p panel", "a ambient", "? help"}

func matches(key string, set []string) bool {
	for _, k := range set {
		if k == key {
			return true
		}
	}
	return false
}
