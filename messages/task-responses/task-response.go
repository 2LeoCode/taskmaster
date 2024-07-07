package task_responses

type TaskResponse interface {
	taskResponseTag()
}

type taskResponse struct{}

func (*taskResponse) taskResponseTag() {}
