package database

type FakeDatabase struct {
	Items  map[PrimaryKey][]byte
	opened bool
}

func NewFakeDatabase() *FakeDatabase {
	return &FakeDatabase{
		Items: make(map[PrimaryKey][]byte),
	}
}

func (f *FakeDatabase) Open(userID string) error {
	_ = userID
	f.opened = true
	return nil
}

func (f *FakeDatabase) Close() error {
	f.opened = false
	return nil
}

func (f *FakeDatabase) ReadValue(pk PrimaryKey) ([]byte, error) {
	f.panicIfNotOpened()
	return f.Items[pk], nil
}

func (f *FakeDatabase) WriteValue(pk PrimaryKey, value []byte) error {
	f.panicIfNotOpened()
	f.Items[pk] = value
	return nil
}

func (f *FakeDatabase) panicIfNotOpened() {
	if !f.opened {
		panic("fake database not opened")
	}
}
