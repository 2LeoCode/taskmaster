package requests

type ReloadConfigRequest interface {
	Request
	reloadConfigTag()
}

type reloadConfigRequest struct{ request }

func (*reloadConfigRequest) reloadConfigTag()

func NewReloadConfigRequest() ReloadConfigRequest {
	return &reloadConfigRequest{}
}
