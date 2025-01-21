package database

type Stub struct {
	OpenErr   error
	CloseErr  error
	ReadErr   error
	WriteErr  error
	QueryErr  error
	DeleteErr error
	Items     map[PrimaryKey][]byte
}

func NewStub() *Stub {
	return &Stub{
		Items: make(map[PrimaryKey][]byte),
	}
}

func (s *Stub) Open(userID string) error {
	return s.OpenErr
}

func (s *Stub) Close() error {
	return s.CloseErr
}

func (s *Stub) ReadValue(pk PrimaryKey) ([]byte, error) {
	if s.ReadErr != nil {
		return nil, s.ReadErr
	}
	return []byte("stub data"), nil
}

func (s *Stub) WriteValue(pk PrimaryKey, value []byte) error {
	return s.WriteErr
}

func (s *Stub) QueryValues(prefix PrimaryKey) (map[string][]byte, error) {
	if s.QueryErr != nil {
		return nil, s.QueryErr
	}
	return map[string][]byte{"stub": []byte("data")}, nil
}

func (s *Stub) DeleteValue(pk PrimaryKey) error {
	return s.DeleteErr
}
