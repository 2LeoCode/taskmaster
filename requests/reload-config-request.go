package requests

type ReloadConfigRequest interface {
	Request
	reloadConfig()
}

type _reloadConfigRequest struct{ _request }

func (*_reloadConfigRequest) reloadConfig() {}

func NewReloadConfigRequest() ReloadConfigRequest {
	return &_reloadConfigRequest{}
}
