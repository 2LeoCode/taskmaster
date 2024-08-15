package helpers

type Issuer interface {
	isIssuer() bool
}

type KnownIssuer interface {
	Issuer
	isKnown() bool
	Id() uint
}

type issuer struct{}

type knownIssuer struct {
	issuer
	id uint
}

func (*issuer) isIssuer() bool { return true }

func (*knownIssuer) isKnown() bool { return true }
func (this *knownIssuer) Id() uint { return this.id }

func newKnownIssuer(id uint) KnownIssuer {
	return &knownIssuer{id: id}
}
