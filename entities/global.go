package entities

type Avatar struct {
	Hash string `json:"hash"`
	URL  string `json:"url"`
}

type Owner struct {
	ID string `json:"id"`
}

// TODO: If we ever decide to add user tokens, we can get the current users rank with ease, potentially modifying the data of entities.Cleanup...
var (
	fakeRank = UserRank{
		Admin:      false,
		Assistant:  false,
		Mod:        false,
		Premium:    false,
		Tester:     false,
		Translator: false,
		Covid:      false,
	}
)
