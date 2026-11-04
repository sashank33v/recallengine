package dto

type AskRequest struct {
	Question string `json:"question"`
}

type AskResponse struct {
	Answer  string            `json:"answer"`
	Sources []PassageResponse `json:"sources"`
}
