package badger

import (
	"context"
	"github.com/dgraph-io/badger/v3"
	"github.com/patrick246/shortlink/pkg/observability/logging"
	"github.com/patrick246/shortlink/pkg/persistence"
	"time"
)

type Repository struct {
	db       *badger.DB
	gcTicker *time.Ticker
}

type Shortlink struct {
	URL string    `json:"url"`
	TTL time.Time `json:"ttl"`
}

var log = logging.CreateLogger("local-storage")

func New(path string) (*Repository, error) {
	db, err := badger.Open(badger.DefaultOptions(path).WithLogger(&badgerLogAdapter{log}))
	if err != nil {
		return nil, err
	}

	ticker := time.NewTicker(5 * time.Minute)
	go func() {
		for range ticker.C {
			for db.RunValueLogGC(0.5) == nil {
			}
		}
	}()

	return &Repository{
		db:       db,
		gcTicker: ticker,
	}, nil
}

func (r *Repository) GetEntryForCode(ctx context.Context, code string) (persistence.Shortlink, error) {
	shortlink := persistence.Shortlink{}
	err := r.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(code))
		if err == badger.ErrKeyNotFound {
			return persistence.ErrNotFound
		}
		if err != nil {
			return err
		}

		ttl := time.Unix(int64(item.ExpiresAt()), 0)

		val, err := item.ValueCopy(nil)
		if err != nil {
			return err
		}
		shortlink = persistence.Shortlink{
			Code: code,
			URL:  string(val),
			TTL:  ttl,
		}
		return nil
	})
	return shortlink, err
}

func (r *Repository) SetEntry(ctx context.Context, shortlink persistence.Shortlink) error {
	return r.db.Update(func(txn *badger.Txn) error {
		entry := badger.NewEntry([]byte(shortlink.Code), []byte(shortlink.URL))
		if !shortlink.TTL.IsZero() {
			entry = entry.WithTTL(shortlink.TTL.Sub(time.Now()))
		}
		return txn.SetEntry(entry)
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
			var ttl time.Time
			if item.ExpiresAt() != 0 {
				ttl = time.Unix(int64(item.ExpiresAt()), 0).UTC()
			}

			url, err := item.ValueCopy(nil)
			if err != nil {
				return err
			}

			sl := persistence.Shortlink{
				Code: string(item.KeyCopy(nil)),
				URL:  string(url),
				TTL:  ttl,
			}
			shortlinks = append(shortlinks, sl)
			i++
		}
		total = i
		return nil
	})
	return shortlinks, total, err
}

func (r *Repository) Close() error {
	r.gcTicker.Stop()
	return r.db.Close()
}
