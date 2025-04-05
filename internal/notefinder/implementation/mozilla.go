package implementation

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

	"notefinder/internal/notefinder/types"
)

/*
Read this for details:
https://www.codejam.info/2021/10/bypass-sqlite-exclusive-lock.html
*/
const (
	bypassExclusiveLock = true
)

type MozillaImplementation struct {
	path string
}

func NewMozillaImplementation(config map[string]string) *MozillaImplementation {
	return &MozillaImplementation{path: config["path"]}
}

func (self *MozillaImplementation) CanWrite() (bool, error) {
	return false, errors.New("Creating new bookmarks is not supported yet")
}

func (self *MozillaImplementation) SupportedProperties() map[string]types.Writable {
	return map[string]types.Writable{"Title": false, "URI": false, "Body": false}
}

func (self *MozillaImplementation) LoadData() (map[uint64]*types.Note, error) {
	data := make(map[uint64]*types.Note, 0)

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

		note := types.NewNote(uint64(id), title)
		note.Set("Body", description, true)
		note.SetFlag(types.FlagReadOnly)
		note.URI = url
		note.Type = types.NoteTypeBookmark

		data[uint64(id)] = note
	}

	return data, err
}

func (self *MozillaImplementation) PutData(note *types.Note) error {
	return errors.New("Creating bookmarks is not currently supported")
}

func (self *MozillaImplementation) UpdateData(oldNote *types.Note, newNote *types.Note) error {
	return errors.New("Editing bookmarks is not currently supported")
}

func (self *MozillaImplementation) DeleteData(note *types.Note) error {
	return errors.New("Deleting bookmarks is not currently supported")
}

func GetMozillaFiles() map[string]string {
	files := make(map[string]string, 0)

	for _, path := range []string{".mozilla/firefox",
		"Library/Application Support/Firefox/Profiles"} {
		user, _ := user.Current()
		baseDir := filepath.Join(user.HomeDir, path)

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
				re := regexp.MustCompile(`/([^/]+)\.default(?:-release)?/`)
				matches := re.FindStringSubmatch(placesFile)

				if len(matches) > 1 {
					log.Println(placesFile)
					files[matches[1]] = placesFile
				}
			}
		} else {
			log.Println(err)
		}
	}

	return files
}
