package database

type Stub struct {
	OpenErr     error
	CloseErr    error
	ReadErr     error
	WriteErr    error
	WriteErrs   map[string]error
	QueryErr    error
	DeleteErr   error
	ReadResult  []byte
	QueryResult map[string][]byte
}

func NewStub() *Stub {
	return &Stub{
		WriteErrs: make(map[string]error),
	}
}

func (s *Stub) Open(userID string) error {
	return s.OpenErr
}

func (s *Stub) Close() error {
	return s.CloseErr
}

func (s *Stub) Read(key string) ([]byte, error) {
	if s.ReadErr != nil {
		return s.ReadResult, s.ReadErr
	}
	return s.ReadResult, nil
}

func (s *Stub) Write(key string, value []byte) error {
	if err, ok := s.WriteErrs[key]; ok {
		return err
	}
	return s.WriteErr
}

func (s *Stub) Query(prefix string) (map[string][]byte, error) {
	if s.QueryErr != nil {
		return s.QueryResult, s.QueryErr
	}
	return s.QueryResult, nil
}

func (s *Stub) Delete(key string) error {
	return s.DeleteErr
}
