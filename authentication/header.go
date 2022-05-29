package authentication

const (
	AuthTokenHeader = "X-Token"
	AuthTokenValue  = "test"
)

type headerAuth interface {
	Login() bool
}

type HeaderAuth struct {
	Name  string
	Value string
}

func (h *HeaderAuth) Login() bool {
	return h.Name == AuthTokenHeader && h.Value == AuthTokenValue
}
