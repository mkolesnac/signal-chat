package apiclient

type Stub struct {
	GetErr     error
	PostErr    error
	GetStatus  int
	PostStatus int
}

func NewStub() *Stub {
	return &Stub{}
}

func (s *Stub) Get(route string, target any) (int, error) {
	return s.GetStatus, s.GetErr
}

func (s *Stub) Post(route string, payload any, target any) (int, error) {
	return s.PostStatus, s.PostErr
}

func (s *Stub) SetAuthorization(username, password string) {
}

func (s *Stub) ClearAuthorization() {
}
