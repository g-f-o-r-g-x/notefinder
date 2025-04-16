package main

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
)

type CommonStorage struct {
	context *Context
	db      *sql.DB
	cache   map[NoteKey]map[string]string
}

var (
	universalStorageRelPath = ".local/share/Notefinder/storage.db"
)

func NewCommonStorage(ctx *Context) *CommonStorage {
	return &CommonStorage{context: ctx}
}

func (self *CommonStorage) Set(note *Note, key string, value *string) {
}

func (self *CommonStorage) Get(note *Note, key string) *string {
	return nil
}
