// handlers.go contains the form/etc handlers needed for beginning a new game.
// Basically, the interface that the "table creator" will access.
// The actual table commands should be all through Websocket.
// This should have an RPC interface.
package wordwalls

import "net/http"

type WordwallsServiceArgs struct {
	// Mode - something like DailyChallenge, NamedList, SavedList, Search
	Mode     string `json:"mode"`
	Filename string `json:"filename"`
	Minimize bool   `json:"minimize"`
}

type WordwallsServiceReply struct {
	Message string `json:"message"`
}

type WordwallsService struct{}

func (w *WordwallsService) NewTable(r *http.Request,
	args *WordwallsServiceArgs, reply *WordwallsServiceReply) error {
	// GenerateGaddag(args.Filename, args.Minimize, true)
	// reply.Message = "Done"
	return nil
}
