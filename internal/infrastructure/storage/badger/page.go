// Package badger stores pages in a badger database. A page is one record: its
// msgpack dump is the on-disk contract, pinned by testdata/page_v1.msgpack.
package badger

import (
	"context"
	"errors"
	"fmt"
	"sort"

	"github.com/dgraph-io/badger/v4"
	"github.com/google/uuid"

	"github.com/derfenix/webarchive/internal/domain"
	"github.com/derfenix/webarchive/internal/infrastructure/storage"
)

func NewPage(db *badger.DB) (*Page, error) {
	return &Page{
		db:     db,
		prefix: []byte("page:"),
	}, nil
}

type Page struct {
	db     *badger.DB
	prefix []byte
}

func (p *Page) GetFile(_ context.Context, pageID, fileID uuid.UUID) (*domain.File, error) {
	page := domain.Page{}
	page.ID = pageID

	var file *domain.File

	err := p.db.View(func(txn *badger.Txn) error {
		data, err := txn.Get(p.key(&page))
		if err != nil {
			return fmt.Errorf("get data: %w", err)
		}

		err = data.Value(func(val []byte) error {
			if err := unmarshal(val, &page); err != nil {
				return fmt.Errorf("unmarshal data: %w", err)
			}

			return nil
		})
		if err != nil {
			return fmt.Errorf("get value: %w", err)
		}

		for i := range page.Results {
			for j := range page.Results[i].Files {
				if page.Results[i].Files[j].ID == fileID {
					file = &page.Results[i].Files[j]
				}
			}
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("view: %w", notFound(err))
	}

	if file == nil {
		return nil, domain.ErrNotFound
	}

	return file, nil
}

func (p *Page) Save(_ context.Context, page *domain.Page) error {
	if p.db.IsClosed() {
		return storage.ErrDBClosed
	}

	marshaled, err := marshal(page)
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

	return nil
}

func (p *Page) Get(_ context.Context, id uuid.UUID) (*domain.Page, error) {
	page := domain.Page{}
	page.ID = id

	err := p.db.View(func(txn *badger.Txn) error {
		data, err := txn.Get(p.key(&page))
		if err != nil {
			return fmt.Errorf("get data: %w", err)
		}

		err = data.Value(func(val []byte) error {
			if err := unmarshal(val, &page); err != nil {
				return fmt.Errorf("unmarshal data: %w", err)
			}

			return nil
		})
		if err != nil {
			return fmt.Errorf("get value: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("view: %w", notFound(err))
	}

	return &page, nil
}

func (p *Page) List(ctx context.Context) ([]*domain.Page, error) {
	pages := make([]*domain.Page, 0, 100)

	err := p.db.View(func(txn *badger.Txn) error {
		iterator := txn.NewIterator(badger.DefaultIteratorOptions)

		defer iterator.Close()

		for iterator.Seek(p.prefix); iterator.ValidForPrefix(p.prefix); iterator.Next() {
			if err := ctx.Err(); err != nil {
				return fmt.Errorf("context canceled: %w", err)
			}

			var page domain.Page

			err := iterator.Item().Value(func(val []byte) error {
				if err := unmarshal(val, &page); err != nil {
					return fmt.Errorf("unmarshal: %w", err)
				}

				return nil
			})
			if err != nil {
				return fmt.Errorf("get item: %w", err)
			}

			pages = append(pages, &page)
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("view: %w", err)
	}

	sort.Slice(pages, func(i, j int) bool {
		return pages[i].Created.After(pages[j].Created)
	})

	return pages, nil
}

func (p *Page) ListPending(ctx context.Context) ([]*domain.Page, error) {
	pages := make([]*domain.Page, 0, 100)

	err := p.db.View(func(txn *badger.Txn) error {
		iterator := txn.NewIterator(badger.DefaultIteratorOptions)

		defer iterator.Close()

		for iterator.Seek(p.prefix); iterator.ValidForPrefix(p.prefix); iterator.Next() {
			if err := ctx.Err(); err != nil {
				return fmt.Errorf("context canceled: %w", err)
			}

			var page domain.Page

			err := iterator.Item().Value(func(val []byte) error {
				if err := unmarshal(val, &page); err != nil {
					return fmt.Errorf("unmarshal: %w", err)
				}

				return nil
			})
			if err != nil {
				return fmt.Errorf("get item: %w", err)
			}

			if page.Status == domain.StatusNew || page.Status == domain.StatusProcessing {
				pages = append(pages, &page)
			}
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("view: %w", err)
	}

	sort.Slice(pages, func(i, j int) bool {
		return pages[i].Created.After(pages[j].Created)
	})

	return pages, nil
}

// notFound maps the storage-level "no such key" to the domain error callers
// switch on.
func notFound(err error) error {
	if errors.Is(err, badger.ErrKeyNotFound) {
		return fmt.Errorf("%w: %s", domain.ErrNotFound, err.Error())
	}

	return err
}

func (p *Page) key(page *domain.Page) []byte {
	return append(p.prefix, []byte(page.ID.String())...)
}
