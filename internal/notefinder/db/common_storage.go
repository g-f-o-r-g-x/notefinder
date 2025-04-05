package db

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"

	"notefinder/internal/notefinder/types"
)

type CommonStorage struct {
	db    *sql.DB
	cache map[types.NoteKey]map[string]string
}

var (
	commonStorageRelPath = ".local/share/Notefinder/storage.db"
)

func NewCommonStorage() *CommonStorage {
	return &CommonStorage{}
}

func (self *CommonStorage) Set(note *types.Note, key string, value *string) {
}

func (self *CommonStorage) Get(note *types.Note, key string) *string {
	return nil
}
