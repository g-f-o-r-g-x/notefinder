package implementation

import (
	_ "encoding/json"
	"errors"

	"notefinder/internal/notefinder/types"
)

type GoogleImplementation struct {
}

func NewGoogleImplementation(config map[string]string) *GoogleImplementation {
	return &GoogleImplementation{}
}

func (self *GoogleImplementation) CanWrite() (bool, error) {
	return false, errors.New("creating new items is not currently supported")
}

func (self *GoogleImplementation) SupportedProperties() map[string]types.Writable {
	return map[string]types.Writable{"Title": false, "URI": false, "Body": false}
}

func (self *GoogleImplementation) LoadData() (map[uint64]*types.Note, error) {
	return make(map[uint64]*types.Note, 0), nil
}

func (self *GoogleImplementation) PutData(note *types.Note) error {
	return errors.New("creating items is not currently supported")
}

func (self *GoogleImplementation) UpdateData(oldNote *types.Note, newNote *types.Note) error {
	return errors.New("editing items is not currently supported")
}

func (self *GoogleImplementation) DeleteData(note *types.Note) error {
	return errors.New("deleting items is not currently supported")
}
