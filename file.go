package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"

	"github.com/gabriel-vasile/mimetype"
)

type FileImplementation struct {
	context      *Context
	path         string
	useExtension bool
}

func NewFileImplementation(ctx *Context, config map[string]string) *FileImplementation {
	return &FileImplementation{context: ctx, path: config["path"],
		useExtension: false}
}

func (self *FileImplementation) CanWrite() (bool, error) {
	tmpFile := "tmpfile"
	file, err := os.CreateTemp(self.path, tmpFile)
	if err != nil {
		return false, err
	}

	defer os.Remove(file.Name())
	defer file.Close()

	return true, nil
}

func (self *FileImplementation) SupportedProperties() map[string]Writable {
	return map[string]Writable{"Title": true, "URI": false, "Body": true}
}

func (self *FileImplementation) LoadData() (map[uint64]*Note, error) {
	mimetype.SetLimit(16) // this was 1024, let's test

	files, err := os.ReadDir(self.path)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	nitems := 0
	for _, f := range files {
		if f.IsDir() {
			continue
		}
		nitems++
	}
	data := make(map[uint64]*Note, nitems)

	for _, f := range files {
		if f.IsDir() {
			continue
		}
		fileName := string(f.Name())

		if regexp.MustCompile(`(^|/)\..*\.sw[pon]$|\.sw[pon]$`).MatchString(fileName) {
			continue
		}

		filePath := filepath.Join(self.path, fileName)
		var stat syscall.Stat_t
		if err := syscall.Stat(filePath, &stat); err != nil {
			log.Println(err)
			continue
		}

		content, err := os.ReadFile(filePath)
		if err != nil {
			log.Println(err)
			continue
		}

		var body string
		if !bytes.ContainsRune(content, 0) {
			body = string(content)
		}
		var setArchived bool
		var name string
		if len(fileName) >= 2 && strings.HasPrefix(fileName, ".") {
			setArchived = true
			name = fileName[1:]
		} else {
			name = fileName
		}

		note := NewNote(self.context, stat.Ino, name)
		note.Set("Body", body, true)

		if setArchived {
			note.SetFlag(FlagArchived)
		}

		if body != "" {
			note.Type = NoteTypeRegular
		} else {
			note.Type = NoteTypeFile
			note.URI = "file://" + filePath

			mime, err := mimetype.DetectFile(filePath)
			if err == nil {
				note.MimeType = mime.String()
			}
		}

		data[stat.Ino] = note
	}

	return data, nil
}

func normalizeTitle(in string) string {
	return strings.ReplaceAll(in, "/", "∕")
}

func (self *FileImplementation) PutData(note *Note) error {
	note.Title = normalizeTitle(note.Title)
	path := filepath.Join(self.path, note.Title)
	_, err := os.Stat(path)
	fmt.Println(note)
	if err == nil {
		err = fmt.Errorf("\"%s\" already exists, cannot create new item", note.Title)
		log.Println(err)
		return err
	}

	if err = os.WriteFile(path, []byte(note.Body), 0644); err != nil {
		log.Println(err)
		return err
	}

	return nil
}

func (self *FileImplementation) UpdateData(oldNote *Note, newNote *Note) error {
	return nil
}

func (self *FileImplementation) DeleteData(note *Note) error {
	return os.Remove(filepath.Join(self.path, filenameString(note)))
}

func filenameString(note *Note) string {
	if !note.FlagIsSet(FlagArchived) {
		return note.Title
	}
	return "." + note.Title
}
