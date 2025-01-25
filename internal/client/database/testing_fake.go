package database

import "strings"

type Fake struct {
	Items  map[PrimaryKey][]byte
	Opened bool
}

func NewFake() *Fake {
	return &Fake{
		Items: make(map[PrimaryKey][]byte),
	}
}

func (f *Fake) Open(userID string) error {
	_ = userID
	f.Opened = true
	return nil
}

func (f *Fake) Close() error {
	f.Opened = false
	return nil
}

func (f *Fake) Read(pk PrimaryKey) ([]byte, error) {
	f.panicIfNotOpened()
	return f.Items[pk], nil
}

func (f *Fake) Write(pk PrimaryKey, value []byte) error {
	f.panicIfNotOpened()
	f.Items[pk] = value
	return nil
}

func (f *Fake) Query(prefix PrimaryKey) (map[string][]byte, error) {
	prefixStr := string(prefix)
	result := make(map[string][]byte)
	for k, v := range f.Items {
		if strings.HasPrefix(string(k), prefixStr) {
			result[string(k)] = v
		}
	}
	return result, nil
}

func (f *Fake) Delete(pk PrimaryKey) error {
	f.panicIfNotOpened()
	delete(f.Items, pk)
	return nil
}

func (f *Fake) panicIfNotOpened() {
	if !f.Opened {
		panic("fake database not Opened")
	}
}
