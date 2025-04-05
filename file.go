package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"syscall"
)

type FileImplementation struct {
	path string
}

func NewFileImplementation(config map[string]string) *FileImplementation {
	return &FileImplementation{path: config["path"]}
}

func (self *FileImplementation) LoadData() (map[uint64]*Note, error) {
	data := make(map[uint64]*Note, 0)

	files, err := os.ReadDir(self.path)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	for _, f := range files {
		if f.IsDir() {
			continue
		}

		filePath := filepath.Join(self.path, f.Name())
		content, err := ioutil.ReadFile(filePath)
		if err != nil {
			log.Println(err)
			continue
		}

		if bytes.ContainsRune(content, 0) {
			continue
		}
		body := string(content)

		var stat syscall.Stat_t
		if err := syscall.Stat(filePath, &stat); err != nil {
			log.Println(err)
			continue
		}
		data[stat.Ino] = NewNote(stat.Ino, f.Name(), body)
		data[stat.Ino].Type = NoteTypeRegular
	}

	return data, nil
}

func (self *FileImplementation) PutData(note *Note) error {
	path := filepath.Join(self.path, note.Title)
	_, err := os.Stat(path)
	if err == nil {
		err = fmt.Errorf(`"%s" already exists, cannot create new item`)
		log.Println(err)
		return err
	}

	if err = ioutil.WriteFile(path, []byte(note.Body), 0644); err != nil {
		log.Println(err)
		return err
	}

	return nil
}

func (self *FileImplementation) UpdateData(oldNote *Note, newNote *Note) error {
	return nil
}

func (self *FileImplementation) DeleteData(note *Note) error {
	return nil
}
