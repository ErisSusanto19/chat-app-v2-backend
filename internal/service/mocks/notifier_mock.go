package mocks

import (
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

type MockNotifier struct {
	mock.Mock
}

func (m *MockNotifier) NotifyUserAddedToGroup(addedByUserID uuid.UUID, newMemberIDs []uuid.UUID, conversationID uuid.UUID) {
	m.Called(addedByUserID, newMemberIDs, conversationID)
}
