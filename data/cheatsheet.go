package data

// Cheatsheet represents a single cheatsheet entry.
type Cheatsheet struct {
	Title       string `json:"title"`
	Slug        string `json:"slug"`
	Description string `json:"description"`
}

type Cheatsheets []Cheatsheet

var cachedCheatsheets Cheatsheets
var cheatsheetsLoaded bool

// GetCheatsheets loads the cheatsheets from the JSON manifest in the order they appear, but only once (cached in memory).
func GetCheatsheets() (Cheatsheets, error) {
	if cheatsheetsLoaded {
		return cachedCheatsheets, nil
	}
	cheatsheets := make(Cheatsheets, 0)
	err := ReadJsonFile("data/cheatsheetsManifest.json", &cheatsheets)
	if err != nil {
		return cheatsheets, err
	}
	cachedCheatsheets = cheatsheets
	cheatsheetsLoaded = true
	return cachedCheatsheets, nil
}

// ClearCheatsheetsCache clears the cached cheatsheets (for testing or reload).
func ClearCheatsheetsCache() {
	cheatsheetsLoaded = false
	cachedCheatsheets = nil
}
