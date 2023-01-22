package boetea

import (
	"context"
	"errors"
)

type Store interface {
	ArtworkStore
	Init(context.Context) error
	Close(context.Context) error
}

var (
	ErrArtworkNotFound = errors.New("artwork not found")
)
