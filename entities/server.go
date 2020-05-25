package entities

type ServerLinks struct {
	Invite   string `json:"invite,omitempty"`
	Website  string `json:"website"`
	Donation string `json:"donation"`
}

type ServerOwner struct {
	ID string `json:"id"`
}

type Server struct {
	ID         string      `bson:"_id" json:"id"`
	InviteCode string      `json:"inviteCode,omitempty"`
	Name       string      `json:"name"`
	ShortDesc  string      `json:"shortDesc"`
	LongDesc   string      `json:"longDesc"`
	Tags       []string    `json:"tags"`
	Owner      ServerOwner `json:"owner"`
	Icon       Avatar      `json:"icon"`
	Links      ServerLinks `json:"links"`
}

// TODO: CleanupServer
func CleanupServer(rank UserRank, server *Server) *Server {
	copied := *server
	copied.InviteCode = ""
	copied.Links.Invite = ""
	if rank.Admin || rank.Assistant {
		copied.InviteCode = server.InviteCode
		copied.Links.Invite = server.Links.Invite
	}
	return server
}
