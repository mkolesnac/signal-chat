package database

import "strings"

type Fake struct {
	Items        map[string][]byte
	Opened       bool
	ActiveUserID string
}

func NewFake() *Fake {
	return &Fake{
		Items: make(map[string][]byte),
	}
}

func (f *Fake) Open(userID string) error {
	f.ActiveUserID = userID
	f.Opened = true
	return nil
}

func (f *Fake) Close() error {
	f.Opened = false
	return nil
}

func (f *Fake) Read(key string) ([]byte, error) {
	f.panicIfNotOpened()
	return f.Items[key], nil
}

func (f *Fake) Write(key string, value []byte) error {
	f.panicIfNotOpened()
	f.Items[key] = value
	return nil
}

func (f *Fake) Query(prefix string) (map[string][]byte, error) {
	prefixStr := prefix
	result := make(map[string][]byte)
	for k, v := range f.Items {
		if strings.HasPrefix(k, prefixStr) {
			result[k] = v
		}
	}
	return result, nil
}

func (f *Fake) Delete(key string) error {
	f.panicIfNotOpened()
	delete(f.Items, key)
	return nil
}

func (f *Fake) panicIfNotOpened() {
	if !f.Opened {
		panic("fake database not Opened")
	}
}
