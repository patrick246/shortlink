package badger

import (
	"context"
	"github.com/dgraph-io/badger/v3"
	"github.com/patrick246/shortlink/pkg/persistence"
)

type Repository struct {
	db *badger.DB
}

func New(path string) (*Repository, error) {
	db, err := badger.Open(badger.DefaultOptions(path))
	if err != nil {
		return nil, err
	}

	return &Repository{
		db: db,
	}, nil
}

func (r *Repository) GetLinkForCode(ctx context.Context, code string) (string, error) {
	url := ""
	err := r.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(code))
		if err == badger.ErrKeyNotFound {
			return persistence.ErrNotFound
		}
		if err != nil {
			return err
		}
		val, err := item.ValueCopy(nil)
		if err != nil {
			return err
		}
		url = string(val)
		return nil
	})
	return url, err
}

func (r *Repository) SetLinkForCode(ctx context.Context, code, url string) error {
	return r.db.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte(code), []byte(url))
	})
}

func (r *Repository) DeleteCode(ctx context.Context, code string) error {
	return r.db.Update(func(txn *badger.Txn) error {
		return txn.Delete([]byte(code))
	})
}

func (r *Repository) GetEntries(ctx context.Context, page, size int64) ([]persistence.Shortlink, int64, error) {
	var shortlinks []persistence.Shortlink
	var total int64

	err := r.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = false
		it := txn.NewIterator(opts)
		defer it.Close()

		skip := page * size
		i := int64(0)
		for it.Rewind(); it.Valid(); it.Next() {
			if i < skip || i >= skip+size {
				i++
				continue
			}

			item := it.Item()
			url, err := item.ValueCopy(nil)
			if err != nil {
				return err
			}
			sl := persistence.Shortlink{
				Code: string(item.KeyCopy(nil)),
				URL:  string(url),
			}
			shortlinks = append(shortlinks, sl)
			i++
		}
		total = i
		return nil
	})
	return shortlinks, total, err
}
