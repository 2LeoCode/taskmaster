package helpers

type Failure interface {
	isFailure() bool
	Reason() string
}

type BaseFailure struct {
	reason string
}

func (*BaseFailure) isFailure() bool              { return true }
func (this *BaseFailure) Reason() string          { return this.reason }
func (this *BaseFailure) SetReason(reason string) { this.reason = reason }
