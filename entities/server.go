package entities

type Server struct {
	ID string `bson:"_id" json:"id"`
}

// TODO: CleanupServer
func CleanupServer(rank UserRank, server *Server) *Server {
	return server
}
