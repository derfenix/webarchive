// Package badgers3 is a placeholder for a repository that keeps page records in
// badger and the files of their results in an S3 bucket. It is not wired into
// the service yet.
package badgers3

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/dgraph-io/badger/v4"
	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"

	"github.com/derfenix/webarchive/internal/domain"
	"github.com/derfenix/webarchive/internal/infrastructure/storage"
)

func NewPage(db *badger.DB, s3 *minio.Client, bucketName string) (*Page, error) {
	return &Page{
		db:         db,
		s3:         s3,
		prefix:     []byte("pages3:"),
		bucketName: bucketName,
	}, nil
}

type Page struct {
	db         *badger.DB
	s3         *minio.Client
	prefix     []byte
	bucketName string
}

func (p *Page) List(ctx context.Context) ([]*domain.Page, error) {
	// TODO implement me
	panic("implement me")
}

func (p *Page) Get(ctx context.Context, id uuid.UUID) (*domain.Page, error) {
	// TODO implement me
	panic("implement me")
}

func (p *Page) GetFile(ctx context.Context, pageID, fileID uuid.UUID) (*domain.File, error) {
	// TODO implement me
	panic("implement me")
}

func (p *Page) Save(ctx context.Context, page *domain.Page) error {
	if p.db.IsClosed() {
		return storage.ErrDBClosed
	}

	marshaled, err := marshal(page.PageBase)
	if err != nil {
		return fmt.Errorf("marshal data: %w", err)
	}

	if err := p.db.Update(func(txn *badger.Txn) error {
		if err := txn.Set(p.key(page), marshaled); err != nil {
			return fmt.Errorf("put data: %w", err)
		}

		return nil
	}); err != nil {
		return fmt.Errorf("update db: %w", err)
	}

	snowball := make(chan minio.SnowballObject, 1)

	go func() {
		defer close(snowball)

		for _, result := range page.Results {
			for _, file := range result.Files {
				for {
					if ctx.Err() != nil {
						return
					}

					if len(snowball) < cap(snowball) {
						break
					}
				}

				snowball <- minio.SnowballObject{
					Key:     file.ID.String(),
					Size:    int64(len(file.Data)),
					ModTime: time.Now(),
					Content: bytes.NewReader(file.Data),
				}
			}
		}
	}()

	if err = p.s3.PutObjectsSnowball(ctx, p.bucketName, minio.SnowballOptions{Compress: true}, snowball); err != nil {
		if dErr := p.db.Update(func(txn *badger.Txn) error {
			if err := txn.Delete(p.key(page)); err != nil {
				return fmt.Errorf("put data: %w", err)
			}

			return nil
		}); dErr != nil {
			err = errors.Join(err, dErr)
		}

		return fmt.Errorf("store files to s3: %w", err)
	}

	return nil
}

func (p *Page) ListPending(ctx context.Context) ([]*domain.Page, error) {
	// TODO implement me
	panic("implement me")
}

func (p *Page) key(page *domain.Page) []byte {
	return append(p.prefix, []byte(page.ID.String())...)
}
