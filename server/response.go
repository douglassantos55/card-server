package server

type Response struct {
	Type    ResponseType
	Payload interface{}
}

type ResponseType string

const (
	Welcome          ResponseType = "welcome"
	WaitForMatch     ResponseType = "wait_for_match"
	MatchFound       ResponseType = "match_found"
	Dequeued         ResponseType = "dequeued"
	WaitOtherPlayers ResponseType = "wait_other_players"
	MatchCanceled    ResponseType = "match_canceled"
)
