package storage

import (
	"context"
	"errors"
)

type Link struct {
	ID       int64
	Link     string
	Username string
}

var (
	ErrEmptyTable    = errors.New("table is empty")
	ErrAlreadyExists = errors.New("Link is already exists")
)

type LinkRepository interface {
	SaveLink(ctx context.Context, link string, username string) (int64, error)
	RandomLink(ctx context.Context, username string) (*Link, error)
	IsExists(ctx context.Context, link string) (int64, error)
}
