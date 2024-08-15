package helpers

type Message interface {
	isMessage() bool
}

type BaseMessage struct{}

func (*BaseMessage) isMessage() bool { return true }
