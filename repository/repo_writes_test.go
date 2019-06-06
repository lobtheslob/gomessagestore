package repository

import (
	"errors"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
)

func TestPostgresRepoWriteMessage(t *testing.T) {
	tests := []struct {
		name        string
		message     *MessageEnvelope
		dbError     error
		expectedErr error
		callCancel  bool
	}{{
		name:        "when there is a db error, return it",
		message:     mockMessages[0],
		dbError:     errors.New("bad things with db happened"),
		expectedErr: errors.New("bad things with db happened"),
	}, {
		name:        "when there is a nil message, an error is returned",
		expectedErr: ErrNilMessage,
	}, {
		name:        "when the message has no ID, an error is returned",
		message:     mockMessageNoID,
		expectedErr: ErrMessageNoID,
	}, {
		name:        "when the message has no stream name, an error is returned",
		message:     mockMessageNoStream,
		expectedErr: ErrInvalidStreamID,
	}, {
		name:    "when there is no db error, it should write the message",
		message: mockMessages[0],
	}, {
		name:       "when it is asked to cancel, it does",
		message:    mockMessages[0],
		callCancel: true,
		dbError:    errors.New("this shouldn't be returned, because we're cancelling"),
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert := assert.New(t)
			db, mockDb, _ := sqlmock.New()
			repo := NewPostgresRepository(db)
			ctx, cancel := context.WithCancel(context.Background())

			if test.message != nil {
				expectedExec := mockDb.
					ExpectExec("SELECT write_message\\(\\$1, \\$2, \\$3, \\$4, \\$5\\)").
					WithArgs(test.message.MessageID,
						test.message.Stream,
						test.message.Type,
						test.message.Data,
						metadataJSON(test.message)).
					WillDelayFor(time.Millisecond * 10)

				if test.dbError == nil {
					expectedExec.WillReturnResult(sqlmock.NewResult(1, 1))
				} else {
					expectedExec.WillReturnError(test.dbError)
				}
			}

			if test.callCancel {
				time.AfterFunc(time.Millisecond*5, cancel) // after the call to the DB, but before it finishes
			}
			err := repo.WriteMessage(ctx, test.message)

			assert.Equal(test.expectedErr, err)
		})
	}
}

func TestPostgresRepoWriteMessageWithExpectedPosition(t *testing.T) {
	tests := []struct {
		name        string
		message     *MessageEnvelope
		dbError     error
		expectedErr error
		position    int64
		callCancel  bool
	}{{
		name:        "when there is a db error, return it",
		message:     mockMessages[0],
		dbError:     errors.New("bad things with db happened"),
		expectedErr: errors.New("bad things with db happened"),
		position:    1,
	}, {
		name:        "when there is a nil message, an error is returned",
		expectedErr: ErrNilMessage,
		position:    1,
	}, {
		name:        "when the message has no ID, an error is returned",
		message:     mockMessageNoID,
		expectedErr: ErrMessageNoID,
		position:    1,
	}, {
		name:        "when the message has no stream name, an error is returned",
		message:     mockMessageNoStream,
		expectedErr: ErrInvalidStreamID,
		position:    1,
	}, {
		name:     "when the position is at 0, no error is returned",
		message:  mockMessages[0],
		position: 0,
	}, {
		name:     "when the position is at -1, no error is returned",
		message:  mockMessages[0],
		position: -1,
	}, {
		name:        "when the position is below -1, an error is returned",
		message:     mockMessages[0],
		expectedErr: ErrInvalidPosition,
		position:    -2,
	}, {
		name:     "when there is no db error, it should write the message",
		message:  mockMessages[0],
		position: 1,
	}, {
		name:       "when it is asked to cancel, it does",
		message:    mockMessages[0],
		position:   0,
		callCancel: true,
		dbError:    errors.New("this shouldn't be returned, because we're cancelling"),
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert := assert.New(t)
			db, mockDb, _ := sqlmock.New()
			repo := NewPostgresRepository(db)
			ctx, cancel := context.WithCancel(context.Background())

			if test.message != nil {
				expectedExec := mockDb.
					ExpectExec("SELECT write_message\\(\\$1, \\$2, \\$3, \\$4, \\$5, \\$6\\)").
					WithArgs(test.message.MessageID,
						test.message.Stream,
						test.message.Type,
						test.message.Data,
						metadataJSON(test.message),
						test.position).
					WillDelayFor(time.Millisecond * 10)

				if test.dbError == nil {
					expectedExec.WillReturnResult(sqlmock.NewResult(1, 1))
				} else {
					expectedExec.WillReturnError(test.dbError)
				}
			}

			if test.callCancel {
				time.AfterFunc(time.Millisecond*5, cancel) // after the call to the DB, but before it finishes
			}
			err := repo.WriteMessageWithExpectedPosition(ctx, test.message, test.position)

			assert.Equal(test.expectedErr, err)
		})
	}
}