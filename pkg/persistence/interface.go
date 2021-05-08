package persistence

import (
	"context"
	"errors"
)

var ErrNotFound = errors.New("not found")

type Shortlink struct {
	Code string
	URL  string
}

type Repository interface {
	GetLinkForCode(ctx context.Context, code string) (string, error)
	SetLinkForCode(ctx context.Context, code, url string) error
	DeleteCode(ctx context.Context, code string) error
	GetEntries(ctx context.Context, page, size int64) ([]Shortlink, int64, error)
	Close() error
}
