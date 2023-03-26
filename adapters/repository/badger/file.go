package badger

import (
	"context"
	"fmt"

	"github.com/dgraph-io/badger/v4"

	"github.com/derfenix/webarchive/entity"
)

func NewFile(db *badger.DB) *File {
	return &File{db: db, prefix: []byte("file:")}
}

type File struct {
	db     *badger.DB
	prefix []byte
}

func (f *File) SaveTx(_ context.Context, txn *badger.Txn, file *entity.File) error {
	if f.db.IsClosed() {
		return ErrDBClosed
	}

	marshaled, err := marshal(file)
	if err != nil {
		return fmt.Errorf("marshal data: %w", err)
	}

	if err := txn.Set(f.key(file), marshaled); err != nil {
		return fmt.Errorf("put data: %w", err)
	}

	return nil
}

func (f *File) key(file *entity.File) []byte {
	return append(f.prefix, []byte(file.ID.String())...)
}
