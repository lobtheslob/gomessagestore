package gomessagestore

import (
	"context"
	"database/sql"

	"github.com/blackhatbrigade/gomessagestore/message"
	"github.com/blackhatbrigade/gomessagestore/projector"
	"github.com/blackhatbrigade/gomessagestore/repository"
)

//go:generate bash -c "${GOPATH}/bin/mockgen github.com/blackhatbrigade/gomessagestore MessageStore > mocks/messagestore.go"

//MessageStore Establishes the interface for Eventide.
type MessageStore interface {
	Write(ctx context.Context, message message.Message, opts ...WriteOption) error
	Get(ctx context.Context, opts ...GetOption) ([]message.Message, error)
	CreateProjector() projector.Projector
}

type msgStore struct {
	repo repository.Repository
}

//GetMessageStoreInterface Grabs a MessageStore instance.
func GetMessageStoreInterface(injectedDB *sql.DB) MessageStore {
	pgRepo := repository.NewPostgresRepository(injectedDB)

	msgstr := &msgStore{
		repo: pgRepo,
	}

	return msgstr
}

//GetMessageStoreInterface2 Grabs a MessageStore instance.
func GetMessageStoreInterface2(injectedRepo repository.Repository) MessageStore {
	msgstr := &msgStore{
		repo: injectedRepo,
	}

	return msgstr
}
