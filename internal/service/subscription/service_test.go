package subscription

import (
	"context"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/neymee/mdexbot/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func init() {
	rand.Seed(time.Now().Unix())
}

type mockCalls []*mock.Call

func (m *mockCalls) Add(c *mock.Call) {
	*m = append(*m, c)
}

func (m *mockCalls) Reset() {
	for _, c := range *m {
		c.Unset()
	}
	*m = mockCalls{}
}

type mdexAPIMock struct {
	mock.Mock
}

func (m *mdexAPIMock) Manga(ctx context.Context, id string) (domain.Manga, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(domain.Manga), args.Error(1)
}

func (m *mdexAPIMock) LastChapters(ctx context.Context, mangaID string, lang *string, publishedSince *time.Time) ([]domain.Chapter, error) {
	args := m.Called(ctx, mangaID, lang, publishedSince)
	return args.Get(0).([]domain.Chapter), args.Error(1)
}

type subRepoMock struct {
	mock.Mock
}

func (m *subRepoMock) UserSubscriptions(ctx context.Context, recipient domain.Recipient) ([]domain.Subscription, error) {
	args := m.Called(ctx, recipient)
	return args.Get(0).([]domain.Subscription), args.Error(1)
}

func (m *subRepoMock) SetUserSubscription(ctx context.Context, recipient domain.Recipient, sub domain.Subscription) error {
	return m.Called(ctx, recipient, sub).Error(0)
}

func (m *subRepoMock) DeleteUserSubscription(ctx context.Context, recipient domain.Recipient, mangaID string, lang string) error {
	return m.Called(ctx, recipient, mangaID, lang).Error(0)
}

func (m *subRepoMock) AllSubscriptions(ctx context.Context) ([]domain.SubscriptionExtended, error) {
	args := m.Called(ctx)
	return args.Get(0).([]domain.SubscriptionExtended), args.Error(1)
}

func (m *subRepoMock) SetSubscriptionLastUpdate(ctx context.Context, sub domain.Subscription, updatedAt time.Time, chapters ...domain.Chapter) error {
	return m.Called(ctx, sub, updatedAt, chapters).Error(0)
}

func (m *subRepoMock) IsChapterNotified(ctx context.Context, sub domain.Subscription, chapter domain.Chapter) (bool, error) {
	args := m.Called(ctx, sub, chapter)
	return args.Bool(0), args.Error(1)
}

func (m *subRepoMock) DeleteAllSubscriptions(ctx context.Context, recipient domain.Recipient) error {
	return m.Called(ctx, recipient).Error(0)
}

func newRecipient() domain.Recipient {
	return domain.RecipientFromInt64(rand.Int63())
}

func TestManga(t *testing.T) {
	expRes1 := domain.Manga{ID: "manga_1"}
	expRes2 := domain.Manga{}

	mdexApi := &mdexAPIMock{}
	mdexApi.On("Manga", mock.Anything, "manga_1").Return(expRes1, nil)
	mdexApi.On("Manga", mock.Anything, "manga_2").Return(expRes2, fmt.Errorf("error"))

	s := New(mdexApi, nil)

	res1, err1 := s.Manga(context.Background(), "manga_1")
	assert.Equal(t, res1, expRes1)
	assert.NoError(t, err1)

	_, err2 := s.Manga(context.Background(), "manga_2")
	assert.Error(t, err2, "error from mdex.Manga expected")

	mdexApi.AssertExpectations(t)
}

func TestList(t *testing.T) {
	rec1, rec2 := newRecipient(), newRecipient()

	expRes1 := []domain.Subscription{{MangaID: "manga_1"}, {MangaID: "manga_2"}}
	expRes2 := ([]domain.Subscription)(nil)

	subRepo := &subRepoMock{}
	subRepo.On("UserSubscriptions", mock.Anything, rec1).Return(expRes1, nil)
	subRepo.On("UserSubscriptions", mock.Anything, rec2).Return(expRes2, fmt.Errorf("error"))

	s := New(nil, subRepo)

	res1, err1 := s.List(context.Background(), rec1)
	assert.NoError(t, err1)
	assert.Equal(t, res1, expRes1)

	_, err2 := s.List(context.Background(), rec2)
	assert.Error(t, err2, "error from storage.UserSubscriptions expected")

	subRepo.AssertExpectations(t)
}

func TestSubscribe_Errors(t *testing.T) {
	user := newRecipient()
	ctx := context.Background()
	sub := domain.Subscription{MangaID: "manga_1", Language: "any", MangaTitle: "manga 1"}

	mdexApi := &mdexAPIMock{}
	subRepo := &subRepoMock{}
	s := New(mdexApi, subRepo)

	calls := mockCalls{}

	// storage.UserSubscriptions error
	calls.Add(subRepo.On("UserSubscriptions", ctx, user).Return([]domain.Subscription{}, fmt.Errorf("error")))
	_, err := s.Subscribe(ctx, user, "", "")
	assert.Error(t, err, "error from storage.UserSubscriptions expected")
	mdexApi.AssertExpectations(t)
	subRepo.AssertExpectations(t)
	calls.Reset()

	// AlreadySubscribedError
	calls.Add(subRepo.On("UserSubscriptions", ctx, user).Return([]domain.Subscription{sub}, nil))
	_, err = s.Subscribe(ctx, user, sub.MangaID, sub.Language)
	assert.Error(t, err, "AlreadySubscribedError error expected")
	assert.Equal(t, err, &AlreadySubscribedError{Manga: sub.MangaTitle, Lang: sub.Language})
	mdexApi.AssertExpectations(t)
	subRepo.AssertExpectations(t)
	calls.Reset()

	// delete any
	calls.Add(subRepo.On("UserSubscriptions", ctx, user).Return([]domain.Subscription{sub}, nil))
	calls.Add(subRepo.On("DeleteUserSubscription", ctx, user, sub.MangaID, sub.Language).Return(fmt.Errorf("error")))
	_, err = s.Subscribe(ctx, user, sub.MangaID, "en")
	assert.Error(t, err, "error from storage.DeleteUserSubscription expected")
	mdexApi.AssertExpectations(t)
	subRepo.AssertExpectations(t)
	calls.Reset()

	// mdex.Manga error
	calls.Add(subRepo.On("UserSubscriptions", ctx, user).Return([]domain.Subscription{}, nil))
	calls.Add(mdexApi.On("Manga", ctx, sub.MangaID).Return(domain.Manga{}, fmt.Errorf("error")))
	_, err = s.Subscribe(ctx, user, sub.MangaID, sub.Language)
	assert.Error(t, err, "error from mdex.Manga expected")
	mdexApi.AssertExpectations(t)
	subRepo.AssertExpectations(t)
	calls.Reset()

	// storage.SetUserSubscription error
	calls.Add(subRepo.On("UserSubscriptions", ctx, user).Return([]domain.Subscription{}, nil))
	calls.Add(mdexApi.On("Manga", ctx, sub.MangaID).Return(domain.Manga{}, nil))
	calls.Add(subRepo.On("SetUserSubscription", ctx, user, mock.Anything).Return(fmt.Errorf("error")))
	_, err = s.Subscribe(ctx, user, sub.MangaID, sub.Language)
	assert.Error(t, err, "error from storage.SetUserSubscription expected")
	mdexApi.AssertExpectations(t)
	subRepo.AssertExpectations(t)
	calls.Reset()
}

func TestSubscribe_Success(t *testing.T) {
	user := newRecipient()
	ctx := context.Background()
	sub1 := domain.Subscription{MangaID: "manga_1", Language: "en", MangaTitle: "manga 1"}
	sub2 := domain.Subscription{MangaID: "manga_1", Language: "es", MangaTitle: "manga 1"}
	expRes := domain.Subscription{MangaID: "manga_1", Language: "any", MangaTitle: "manga 1"}

	mdexApi := &mdexAPIMock{}
	subRepo := &subRepoMock{}
	s := New(mdexApi, subRepo)

	subRepo.On("UserSubscriptions", ctx, user).Return([]domain.Subscription{sub1, sub2}, nil)
	subRepo.On("DeleteUserSubscription", ctx, user, sub1.MangaID, sub1.Language).Return(nil)
	subRepo.On("DeleteUserSubscription", ctx, user, sub2.MangaID, sub2.Language).Return(nil)
	mdexApi.On("Manga", ctx, expRes.MangaID).Return(
		domain.Manga{ID: expRes.MangaID, Title: map[string]string{expRes.Language: expRes.MangaTitle}},
		nil,
	)
	subRepo.On("SetUserSubscription", ctx, user, expRes).Return(nil)

	res, err := s.Subscribe(ctx, user, expRes.MangaID, expRes.Language)
	assert.NoError(t, err)
	assert.Equal(t, expRes, res)
	mdexApi.AssertExpectations(t)
	subRepo.AssertExpectations(t)
}

func TestUnsubscribe_Errors(t *testing.T) {
	user := newRecipient()
	ctx := context.Background()

	subRepo := &subRepoMock{}
	s := New(nil, subRepo)

	sub1 := domain.Subscription{MangaID: "manga_1", Language: "en"}

	calls := mockCalls{}

	// storage.UserSubscriptions error
	calls.Add(subRepo.On("UserSubscriptions", ctx, user).Return(([]domain.Subscription)(nil), fmt.Errorf("error")))
	_, err := s.Unsubscribe(ctx, user, "", "")
	assert.Error(t, err, "error from storage.UserSubscriptions expected")
	subRepo.AssertExpectations(t)
	calls.Reset()

	// ErrNoSuchSubscription
	calls.Add(subRepo.On("UserSubscriptions", ctx, user).Return([]domain.Subscription{sub1}, nil))
	_, err = s.Unsubscribe(ctx, user, "manga_2", "en")
	assert.ErrorIs(t, err, ErrNoSuchSubscription, "ErrNoSuchSubscription error expected")
	subRepo.AssertExpectations(t)
	calls.Reset()

	// storage.DeleteUserSubscription errr
	calls.Add(subRepo.On("UserSubscriptions", ctx, user).Return([]domain.Subscription{sub1}, nil))
	calls.Add(subRepo.On("DeleteUserSubscription", ctx, user, sub1.MangaID, sub1.Language).Return(fmt.Errorf("error")))
	_, err = s.Unsubscribe(ctx, user, sub1.MangaID, sub1.Language)
	assert.Error(t, err, "error from storage.DeleteUserSubscription expected")
	subRepo.AssertExpectations(t)
}

func TestUnsubscribe_Success(t *testing.T) {
	user := newRecipient()
	ctx := context.Background()

	subRepo := &subRepoMock{}
	s := New(nil, subRepo)

	sub1 := domain.Subscription{MangaID: "manga_1", Language: "en"}
	sub2 := domain.Subscription{MangaID: "manga_1", Language: "es"}

	// storage.DeleteUserSubscription errr
	subRepo.On("UserSubscriptions", ctx, user).Return([]domain.Subscription{sub1, sub2}, nil)
	subRepo.On("DeleteUserSubscription", ctx, user, sub1.MangaID, sub1.Language).Return(nil)
	resSub, err := s.Unsubscribe(ctx, user, sub1.MangaID, sub1.Language)
	assert.NoError(t, err)
	assert.Equal(t, sub1, resSub)
	subRepo.AssertExpectations(t)
}

func TestUnsubscribeAll(t *testing.T) {
	rec1, rec2 := newRecipient(), newRecipient()
	ctx := context.Background()

	subRepo := &subRepoMock{}
	subRepo.On("DeleteAllSubscriptions", ctx, rec1).Return(nil)
	subRepo.On("DeleteAllSubscriptions", ctx, rec2).Return(fmt.Errorf("error"))

	s := New(nil, subRepo)

	err1 := s.UnsubscribeAll(ctx, rec1)
	assert.NoError(t, err1)

	err2 := s.UnsubscribeAll(ctx, rec2)
	assert.Error(t, err2, "error from storage.DeleteAllSubscriptions expected")

	subRepo.AssertExpectations(t)
}

func TestUpdates_Error(t *testing.T) {
	user := newRecipient()
	ctx := context.Background()

	sub1 := domain.SubscriptionExtended{
		Subscription: domain.Subscription{MangaID: "manga_1", Language: "en"},
		Recipients:   []domain.Recipient{user},
		UpdatedAt:    time.Date(2022, 12, 1, 0, 0, 0, 0, time.Local),
	}

	chap1 := domain.Chapter{ID: "ch_1"}

	mdexApi := &mdexAPIMock{}
	subRepo := &subRepoMock{}
	s := New(mdexApi, subRepo)

	calls := mockCalls{}

	// storage.AllSubscriptions error
	calls.Add(subRepo.On("AllSubscriptions", ctx).Return(([]domain.SubscriptionExtended)(nil), fmt.Errorf("error")))
	_, err := s.Updates(ctx)
	assert.Error(t, err, "error from storage.AllSubscriptions expected")
	subRepo.AssertExpectations(t)
	mdexApi.AssertExpectations(t)
	calls.Reset()

	// cancelled context error
	cancelledCtx, cancel := context.WithCancel(ctx)
	cancel()
	calls.Add(subRepo.On("AllSubscriptions", cancelledCtx).Return([]domain.SubscriptionExtended{sub1}, nil))
	_, err = s.Updates(cancelledCtx)
	assert.Error(t, err, "error caused by cancelled context expected")
	subRepo.AssertExpectations(t)
	mdexApi.AssertExpectations(t)
	calls.Reset()

	// mdex.LastChapters error
	calls.Add(subRepo.On("AllSubscriptions", ctx).Return([]domain.SubscriptionExtended{sub1}, nil))
	calls.Add(mdexApi.On("LastChapters", ctx, sub1.MangaID, &sub1.Language, &sub1.UpdatedAt).Return(([]domain.Chapter)(nil), fmt.Errorf("error")))
	_, err = s.Updates(ctx)
	assert.Error(t, err, "error mdex.LastChapters expected")
	subRepo.AssertExpectations(t)
	mdexApi.AssertExpectations(t)
	calls.Reset()

	// storage.IsChapterNotified error
	calls.Add(subRepo.On("AllSubscriptions", ctx).Return([]domain.SubscriptionExtended{sub1}, nil))
	calls.Add(mdexApi.On("LastChapters", ctx, sub1.MangaID, &sub1.Language, &sub1.UpdatedAt).Return([]domain.Chapter{chap1}, nil))
	calls.Add(subRepo.On("IsChapterNotified", ctx, sub1.Subscription, chap1).Return(false, fmt.Errorf("error")))
	_, err = s.Updates(ctx)
	assert.Error(t, err, "error storage.IsChapterNotified expected")
	subRepo.AssertExpectations(t)
	mdexApi.AssertExpectations(t)
	calls.Reset()

	// storage.SetSubscriptionLastUpdate error
	calls.Add(subRepo.On("AllSubscriptions", ctx).Return([]domain.SubscriptionExtended{sub1}, nil))
	calls.Add(mdexApi.On("LastChapters", ctx, sub1.MangaID, &sub1.Language, &sub1.UpdatedAt).Return([]domain.Chapter{}, nil))
	calls.Add(subRepo.On("SetSubscriptionLastUpdate", ctx, sub1.Subscription, mock.Anything, []domain.Chapter{}).Return(fmt.Errorf("error")))
	_, err = s.Updates(ctx)
	assert.Error(t, err, "error storage.SetSubscriptionLastUpdate expected")
	subRepo.AssertExpectations(t)
	mdexApi.AssertExpectations(t)
	calls.Reset()
}

func TestUpdates_Success(t *testing.T) {
	user := newRecipient()
	ctx := context.Background()

	sub1 := domain.SubscriptionExtended{
		Subscription: domain.Subscription{MangaID: "manga_1", Language: "any", MangaTitle: "manga 1"},
		Recipients:   []domain.Recipient{user},
		UpdatedAt:    time.Date(2022, 12, 1, 0, 0, 0, 0, time.Local),
	}
	chap1 := domain.Chapter{ID: "ch_1", Volume: "1", Chapter: "1", Title: "chap1"}
	chap1dup := domain.Chapter{ID: "ch_1", Volume: "1", Chapter: "1", Title: "chap1 dup"} // should be filtered

	expUpdates := []domain.Update{
		{
			MangaID:     sub1.MangaID,
			MangaTitle:  sub1.MangaTitle,
			Language:    sub1.Language,
			NewChapters: []domain.Chapter{chap1},
			Recipients:  sub1.Recipients,
		},
	}

	mdexApi := &mdexAPIMock{}
	subRepo := &subRepoMock{}
	s := New(mdexApi, subRepo)

	subRepo.On("AllSubscriptions", ctx).Return([]domain.SubscriptionExtended{sub1}, nil)
	mdexApi.On("LastChapters", ctx, sub1.MangaID, (*string)(nil), &sub1.UpdatedAt).Return([]domain.Chapter{chap1, chap1dup}, nil)
	subRepo.On("IsChapterNotified", ctx, sub1.Subscription, chap1).Return(false, nil)
	subRepo.On("IsChapterNotified", ctx, sub1.Subscription, chap1dup).Return(false, nil)
	subRepo.On("SetSubscriptionLastUpdate", ctx, sub1.Subscription, mock.Anything, []domain.Chapter{chap1}).Return(nil)
	updates, err := s.Updates(ctx)
	assert.NoError(t, err)
	assert.ElementsMatch(t, updates, expUpdates)
	subRepo.AssertExpectations(t)
	mdexApi.AssertExpectations(t)
}
