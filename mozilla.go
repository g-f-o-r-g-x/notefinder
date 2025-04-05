package main

import (
	"database/sql"
	"errors"
	"io/ioutil"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

type MozillaImplementation struct {
	path string
}

func NewMozillaImplementation(config map[string]string) *MozillaImplementation {
	return &MozillaImplementation{path: config["path"]}
}

func (self *MozillaImplementation) LoadData() (map[uint64]*Note, error) {
	data := make(map[uint64]*Note, 0)

	file, err := ioutil.TempFile("/tmp/", "nf.*.sqlite")
	if err != nil {
		log.Fatal(err)
	}
	defer os.Remove(file.Name())

	bytes, err := ioutil.ReadFile(self.path)
	err = ioutil.WriteFile(file.Name(), bytes, 0644)
	db, err := sql.Open("sqlite3", file.Name())

	query := "select b.id, b.title, p.url from moz_bookmarks b, moz_places p where b.fk = p.id"
	rows, _ := db.Query(query)
	for rows.Next() {
		var id int
		var title string
		var url string
		err = rows.Scan(&id, &title, &url)
		if err != nil {
			log.Fatal(err)
		}

		data[uint64(id)] = NewNote(uint64(id), title, url)
		data[uint64(id)].Type = NoteTypeBookmark
	}

	return data, err
}

func (self *MozillaImplementation) PutData(note *Note) error {
	return errors.New("Creating bookmarks is not currently supported")
}

func (self *MozillaImplementation) UpdateData(oldNote *Note, newNote *Note) error {
	return errors.New("Editing bookmarks is not currently supported")
}

func (self *MozillaImplementation) DeleteData(note *Note) error {
	return errors.New("Deleting bookmarks is not currently supported")
}

func getMozillaFiles() []string {
	files := make([]string, 0)

	user, _ := user.Current()
	baseDir := filepath.Join(user.HomeDir, ".mozilla/firefox")

	items, err := os.ReadDir(baseDir)
	if err == nil {
		for _, f := range items {
			if !f.IsDir() || !strings.Contains(f.Name(), "default") {
				continue
			}

			placesFile := filepath.Join(baseDir, f.Name(), "places.sqlite")
			_, err := os.Stat(placesFile)
			if err != nil {
				log.Println(err)
				continue
			}

			files = append(files, placesFile)
		}
	} else {
		log.Println(err)
	}

	return files
}
