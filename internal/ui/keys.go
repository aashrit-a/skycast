package ui

// Key names as reported by tea.KeyMsg.String().
var (
	keysQuit    = []string{"q", "esc", "ctrl+c"}
	keysRefresh = []string{"r"}
	keysUnits   = []string{"u"}
	keysPanel   = []string{"p"}
	keysAmbient = []string{"a"}
	keysHelp    = []string{"?"}
)

// footerHelp is the one-line key reference shown in the footer.
const footerHelp = "q quit · r refresh · u units · p panel · a ambient · ? help"

func matches(key string, set []string) bool {
	for _, k := range set {
		if k == key {
			return true
		}
	}
	return false
}
