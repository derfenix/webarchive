package badger

import (
	"context"
	"fmt"
	"sort"

	"github.com/dgraph-io/badger/v4"
	"github.com/google/uuid"

	"github.com/derfenix/webarchive/entity"
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

func (p *Page) GetFile(_ context.Context, pageID, fileID uuid.UUID) (*entity.File, error) {
	page := entity.Page{ID: pageID}
	var file *entity.File

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
				ff := &page.Results[i].Files[j]

				if ff.ID == fileID {
					file = ff
				}
			}

		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("view: %w", err)
	}

	return file, nil
}

func (p *Page) Save(_ context.Context, page *entity.Page) error {
	if p.db.IsClosed() {
		return ErrDBClosed
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

func (p *Page) Get(_ context.Context, id uuid.UUID) (*entity.Page, error) {
	site := entity.Page{ID: id}

	err := p.db.View(func(txn *badger.Txn) error {
		data, err := txn.Get(p.key(&site))
		if err != nil {
			return fmt.Errorf("get data: %w", err)
		}

		err = data.Value(func(val []byte) error {
			if err := unmarshal(val, &site); err != nil {
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
		return nil, fmt.Errorf("view: %w", err)
	}

	return &site, nil
}

func (p *Page) ListAll(ctx context.Context) ([]*entity.Page, error) {
	pages := make([]*entity.Page, 0, 100)

	err := p.db.View(func(txn *badger.Txn) error {
		iterator := txn.NewIterator(badger.DefaultIteratorOptions)

		defer iterator.Close()

		for iterator.Seek(p.prefix); iterator.ValidForPrefix(p.prefix); iterator.Next() {
			if err := ctx.Err(); err != nil {
				return fmt.Errorf("context canceled: %w", err)
			}

			var page entity.Page

			err := iterator.Item().Value(func(val []byte) error {
				if err := unmarshal(val, &page); err != nil {
					return fmt.Errorf("unmarshal: %w", err)
				}

				return nil
			})

			if err != nil {
				return fmt.Errorf("get item: %w", err)
			}

			pages = append(pages, &entity.Page{
				ID:          page.ID,
				URL:         page.URL,
				Description: page.Description,
				Created:     page.Created,
				Formats:     page.Formats,
				Version:     page.Version,
				Status:      page.Status,
				Meta:        page.Meta,
			})
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

func (p *Page) ListUnprocessed(ctx context.Context) ([]entity.Page, error) {
	pages := make([]entity.Page, 0, 100)

	err := p.db.View(func(txn *badger.Txn) error {
		iterator := txn.NewIterator(badger.DefaultIteratorOptions)

		defer iterator.Close()

		for iterator.Seek(p.prefix); iterator.ValidForPrefix(p.prefix); iterator.Next() {
			if err := ctx.Err(); err != nil {
				return fmt.Errorf("context canceled: %w", err)
			}

			var page entity.Page

			err := iterator.Item().Value(func(val []byte) error {
				if err := unmarshal(val, &page); err != nil {
					return fmt.Errorf("unmarshal: %w", err)
				}

				return nil
			})

			if err != nil {
				return fmt.Errorf("get item: %w", err)
			}

			if page.Status != entity.StatusProcessing {
				continue
			}

			//goland:noinspection GoVetCopyLock
			pages = append(pages, page) //nolint:govet // didn't touch the lock here
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

func (p *Page) key(site *entity.Page) []byte {
	return append(p.prefix, []byte(site.ID.String())...)
}
