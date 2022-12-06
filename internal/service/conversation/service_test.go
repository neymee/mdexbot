package conversation

import (
	"context"
	"fmt"
	"math/rand"
	"testing"

	"github.com/neymee/mdexbot/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type convRepoMock struct {
	mock.Mock
}

func (r *convRepoMock) ConversationContext(ctx context.Context, recipient domain.Recipient) (string, error) {
	args := r.Called(ctx, recipient)
	return args.String(0), args.Error(1)
}

func (r *convRepoMock) SetConversationContext(ctx context.Context, recipient domain.Recipient, cmd string) error {
	return r.Called(ctx, recipient, cmd).Error(0)
}

func (r *convRepoMock) DeleteConversationContext(ctx context.Context, recipient domain.Recipient) error {
	return r.Called(ctx, recipient).Error(0)
}

func newRecipient() domain.Recipient {
	return domain.RecipientFromInt64(rand.Int63())
}

func TestConversationContext(t *testing.T) {
	r := &convRepoMock{}
	s := New(r)

	ctx := context.Background()
	user1, user2 := newRecipient(), newRecipient()

	r.On("ConversationContext", ctx, user1).Return("test", nil)
	r.On("ConversationContext", ctx, user2).Return("", fmt.Errorf("error"))

	conv, err1 := s.ConversationContext(ctx, user1)
	assert.NoError(t, err1)
	assert.Equal(t, "test", conv)

	_, err2 := s.ConversationContext(ctx, user2)
	assert.Error(t, err2, "error from repo.ConversationContext expected")
}

func TestSetConversationContext(t *testing.T) {
	r := &convRepoMock{}
	s := New(r)

	ctx := context.Background()
	user := newRecipient()

	r.On("SetConversationContext", ctx, user, "test1").Return(nil)
	r.On("SetConversationContext", ctx, user, "test2").Return(fmt.Errorf("error"))

	err1 := s.SetConversationContext(ctx, user, "test1")
	assert.NoError(t, err1)

	err2 := s.SetConversationContext(ctx, user, "test2")
	assert.Error(t, err2, "error from repo.SetConversationContext expected")
}

func TestDeleteConversationContext(t *testing.T) {
	r := &convRepoMock{}
	s := New(r)

	ctx := context.Background()
	user1, user2 := newRecipient(), newRecipient()

	r.On("DeleteConversationContext", ctx, user1).Return(nil)
	r.On("DeleteConversationContext", ctx, user2).Return(fmt.Errorf("error"))

	err1 := s.DeleteConversationContext(ctx, user1)
	assert.NoError(t, err1)

	err2 := s.DeleteConversationContext(ctx, user2)
	assert.Error(t, err2, "error from repo.DeleteConversationContext expected")
}
