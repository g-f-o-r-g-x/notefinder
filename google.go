package main

import (
	_ "encoding/json"
	"errors"
)

type GoogleImplementation struct {
	context *Context
}

func NewGoogleImplementation(ctx *Context, config map[string]string) *GoogleImplementation {
	return &GoogleImplementation{context: ctx}
}

func (self *GoogleImplementation) CanWrite() (bool, error) {
	return false, errors.New("creating new items is not currently supported")
}

func (self *GoogleImplementation) SupportedProperties() map[string]Writable {
	return map[string]Writable{"Title": false, "URI": false, "Body": false}
}

func (self *GoogleImplementation) LoadData() (map[uint64]*Note, error) {
	return make(map[uint64]*Note, 0), nil
}

func (self *GoogleImplementation) PutData(note *Note) error {
	return errors.New("creating items is not currently supported")
}

func (self *GoogleImplementation) UpdateData(oldNote *Note, newNote *Note) error {
	return errors.New("editing items is not currently supported")
}

func (self *GoogleImplementation) DeleteData(note *Note) error {
	return errors.New("deleting items is not currently supported")
}
