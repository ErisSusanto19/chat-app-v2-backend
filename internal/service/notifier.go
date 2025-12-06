package service

import "github.com/google/uuid"

type Notifier interface {
	NotifyUserAddedToGroup(addedByUserID uuid.UUID, newMemberIDs []uuid.UUID, conversationID uuid.UUID)
}
