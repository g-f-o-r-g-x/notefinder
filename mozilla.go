package main

import (
	"database/sql"
	"errors"
	"io/ioutil"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"regexp"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

/*
Read this for details:
https://www.codejam.info/2021/10/bypass-sqlite-exclusive-lock.html
*/
const (
	bypassExclusiveLock = true
)

type MozillaImplementation struct {
	context *Context
	path    string
}

func NewMozillaImplementation(ctx *Context, config map[string]string) *MozillaImplementation {
	return &MozillaImplementation{context: ctx, path: config["path"]}
}

func (self *MozillaImplementation) CanWrite() (bool, error) {
	return false, errors.New("Creating new bookmarks is not supported yet")
}

func (self *MozillaImplementation) LoadData() (map[uint64]*Note, error) {
	data := make(map[uint64]*Note, 0)

	var fileName string
	if !bypassExclusiveLock {
		file, err := ioutil.TempFile("/tmp/", "nf.*.sqlite")
		if err != nil {
			log.Fatal(err)
		}
		defer os.Remove(file.Name())

		bytes, err := ioutil.ReadFile(self.path)
		err = ioutil.WriteFile(file.Name(), bytes, 0644)

		fileName = file.Name()
	} else {
		fileName = "file:" + self.path + "?immutable=1"
	}
	db, err := sql.Open("sqlite3", fileName)

	query := `select b.id, b.title, p.url, ifnull(p.description, "")
		from moz_bookmarks b, moz_places p where b.fk = p.id`
	rows, _ := db.Query(query)
	for rows.Next() {
		var id int
		var title string
		var url string
		var description string
		err = rows.Scan(&id, &title, &url, &description)
		if err != nil {
			log.Fatal(err)
		}

		data[uint64(id)] = NewNote(self.context, uint64(id), title, description)
		data[uint64(id)].URI = url
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

func getMozillaFiles() map[string]string {
	files := make(map[string]string, 0)

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
			re := regexp.MustCompile(`/firefox/([^/]+)\.default(?:-release)?/`)
			matches := re.FindStringSubmatch(placesFile)

			if len(matches) > 1 {
				files[matches[1]] = placesFile
			}
		}
	} else {
		log.Println(err)
	}

	return files
}
