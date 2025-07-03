package implementation

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

	"notefinder/internal/notefinder/types"
)

type FileImplementation struct {
	path         string
	useExtension bool
}

func NewFileImplementation(config map[string]string) *FileImplementation {
	return &FileImplementation{path: config["path"],
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

func (self *FileImplementation) SupportedProperties() map[string]types.Writable {
	return map[string]types.Writable{"Title": true, "URI": false, "Body": true}
}

func processDir(path string, dst map[uint64]*types.Note, paths []string) error {
	files, err := os.ReadDir(path)
	if err != nil {
		log.Println(err)
		return err
	}

	for _, f := range files {
		if f.IsDir() {
			paths = append(paths, string(f.Name()))
			err = processDir(filepath.Join(path, string(f.Name())), dst, paths)
			if err != nil {
				log.Println(err)
			}
			continue
		}
		fileName := string(f.Name())

		// vim temporary files
		if regexp.MustCompile(`(^|/)\..*\.sw[pon]$|\.sw[pon]$`).MatchString(fileName) {
			continue
		}

		filePath := filepath.Join(path, fileName)
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

		note := types.NewNote(stat.Ino, name)
		note.Set("Body", body, true)

		if setArchived {
			note.SetFlag(types.FlagArchived)
		}

		if body != "" {
			note.Type = types.NoteTypeRegular
		} else {
			note.Type = types.NoteTypeFile
			note.URI = "file://" + filePath

			mime, err := mimetype.DetectFile(filePath)
			if err == nil {
				note.MimeType = mime.String()
			}
		}
		if len(paths) > 0 {
			note.Tags = make([]string, len(paths))
			copy(note.Tags, paths)
		}

		dst[stat.Ino] = note
	}

	return nil
}

func (self *FileImplementation) LoadData() (map[uint64]*types.Note, error) {
	mimetype.SetLimit(16) // this was 1024, let's test
	data := make(map[uint64]*types.Note, 0)
	paths := make([]string, 0)
	if err := processDir(self.path, data, paths); err != nil {
		return nil, err
	}

	return data, nil
}

func normalizeTitle(in string) string {
	return strings.ReplaceAll(in, "/", "∕")
}

func (self *FileImplementation) PutData(note *types.Note) error {
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

func (self *FileImplementation) UpdateData(oldNote *types.Note, newNote *types.Note) error {
	return nil
}

func (self *FileImplementation) DeleteData(note *types.Note) error {
	return os.Remove(filepath.Join(self.path, filenameString(note)))
}

func filenameString(note *types.Note) string {
	if !note.FlagIsSet(types.FlagArchived) {
		return note.Title
	}
	return "." + note.Title
}
