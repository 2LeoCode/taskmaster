package process_responses

type ProcessResponse interface {
	processResponseTag()
}

type processResponse struct{}

func (*processResponse) processResponseTag()
