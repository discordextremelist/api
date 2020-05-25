package entities

type BotStatus struct {
	Approved bool `json:"approved"`
	Premium  bool `json:"premium,omitempty"`
	SiteBot  bool `json:"siteBot"`
	Archived bool `json:"archived"`
}

type BotVotes struct {
	Positive []string `json:"positive"`
	Negative []string `json:"negative"`
}

type BotOwner struct {
	ID string `json:"id"`
}

type BotLinks struct {
	Invite   string `json:"invite"`
	Support  string `json:"support"`
	Website  string `json:"website"`
	Donation string `json:"donation"`
	Repo     string `json:"repo"`
}

type WidgetBot struct {
	Channel string `json:"channel"`
	Options string `json:"options"`
	Server  string `json:"server"`
}

type Bot struct {
	ID          string    `bson:"_id" json:"id"`
	Name        string    `json:"name"`
	Prefix      string    `json:"prefix"`
	Tags        []string  `json:"tags"`
	VanityURL   string    `json:"vanityUrl"`
	ServerCount int       `json:"serverCount"`
	ShardCount  int       `json:"shardCount"`
	Token       string    `json:"token,omitempty"`
	ShortDesc   string    `json:"shortDesc"`
	LongDesc    string    `json:"longDesc"`
	ModNotes    string    `json:"modNotes,omitempty"`
	Editors     []string  `json:"editors"`
	Owner       BotOwner  `json:"owner"`
	Avatar      Avatar    `json:"avatar"`
	Votes       *BotVotes `json:"votes,omitempty"`
	Links       BotLinks  `json:"links"`
	Status      BotStatus `json:"status"`
}

func CleanupBot(rank UserRank, bot *Bot) *Bot {
	copied := *bot
	copied.ModNotes = ""
	copied.Token = ""
	copied.Votes = nil
	copied.Status.Premium = false
	if rank.Mod {
		copied.ModNotes = bot.ModNotes
	}
	if rank.Admin || rank.Assistant {
		copied.ModNotes = bot.ModNotes
		copied.Token = bot.Token
		copied.Votes = bot.Votes
		copied.Status.Premium = bot.Status.Premium
	}
	return &copied
}
