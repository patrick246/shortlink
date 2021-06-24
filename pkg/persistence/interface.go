package persistence

import (
	"context"
	"errors"
	"time"
)

var ErrNotFound = errors.New("not found")

type Shortlink struct {
	Code string
	URL  string
	TTL  time.Time
}

type Repository interface {
	GetEntryForCode(ctx context.Context, code string) (Shortlink, error)
	SetEntry(ctx context.Context, shortlink Shortlink) error
	DeleteCode(ctx context.Context, code string) error
	GetEntries(ctx context.Context, page, size int64) ([]Shortlink, int64, error)
	Close() error
}
