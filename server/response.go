package server

type Response struct {
	Type ResponseType
}

type ResponseType string

const (
	Welcome      ResponseType = "welcome"
	WaitForMatch ResponseType = "wait_for_match"
	MatchFound   ResponseType = "match_found"
)
