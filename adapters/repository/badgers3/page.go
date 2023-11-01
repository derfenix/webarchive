package badgers3

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/derfenix/webarchive/adapters/repository"
	"github.com/derfenix/webarchive/entity"
	"github.com/dgraph-io/badger/v4"
	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
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

func (p *Page) ListAll(ctx context.Context) ([]*entity.PageBase, error) {
	// TODO implement me
	panic("implement me")
}

func (p *Page) Get(ctx context.Context, id uuid.UUID) (*entity.Page, error) {
	// TODO implement me
	panic("implement me")
}

func (p *Page) GetFile(ctx context.Context, pageID, fileID uuid.UUID) (*entity.File, error) {
	// TODO implement me
	panic("implement me")
}

func (p *Page) Save(ctx context.Context, page *entity.Page) error {
	if p.db.IsClosed() {
		return repository.ErrDBClosed
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

func (p *Page) ListUnprocessed(ctx context.Context) ([]entity.Page, error) {
	// TODO implement me
	panic("implement me")
}

func (p *Page) key(site *entity.Page) []byte {
	return append(p.prefix, []byte(site.ID.String())...)
}
