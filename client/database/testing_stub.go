package database

type Stub struct {
	OpenErr     error
	CloseErr    error
	ReadErr     error
	WriteErr    error
	WriteErrs   map[PrimaryKey]error
	QueryErr    error
	DeleteErr   error
	ReadResult  []byte
	QueryResult map[string][]byte
}

func NewStub() *Stub {
	return &Stub{
		WriteErrs: make(map[PrimaryKey]error),
	}
}

func (s *Stub) Open(userID string) error {
	return s.OpenErr
}

func (s *Stub) Close() error {
	return s.CloseErr
}

func (s *Stub) Read(pk PrimaryKey) ([]byte, error) {
	if s.ReadErr != nil {
		return s.ReadResult, s.ReadErr
	}
	return s.ReadResult, nil
}

func (s *Stub) Write(pk PrimaryKey, value []byte) error {
	if err, ok := s.WriteErrs[pk]; ok {
		return err
	}
	return s.WriteErr
}

func (s *Stub) Query(prefix PrimaryKey) (map[string][]byte, error) {
	if s.QueryErr != nil {
		return s.QueryResult, s.QueryErr
	}
	return s.QueryResult, nil
}

func (s *Stub) Delete(pk PrimaryKey) error {
	return s.DeleteErr
}
