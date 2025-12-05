package service

import (
	"context"
	"errors"
	"strings"

	"github.com/ErisSusanto19/chat-app-v2-backend/internal/domain"
	"github.com/ErisSusanto19/chat-app-v2-backend/internal/repository"
	"github.com/google/uuid"
)

type ContactService interface {
	AddContact(ctx context.Context, ownerID uuid.UUID, aliasName, contactEmail string) (*domain.Contact, error)
	GetContacts(ctx context.Context, ownerID uuid.UUID) ([]*domain.Contact, error)
	UpdateContact(ctx context.Context, contactID, ownerID uuid.UUID, newAliasName string) error
	DeleteContact(ctx context.Context, contactID, ownerID uuid.UUID) error
	GetContactDetail(ctx context.Context, contactID, ownerID uuid.UUID) (*repository.ContactDetail, error)
}

type contactService struct {
	contactRepo repository.ContactRepository
	userRepo    repository.UserRepository
}

func NewContactService(contactRepo repository.ContactRepository, userRepo repository.UserRepository) ContactService {
	return &contactService{
		contactRepo: contactRepo,
		userRepo:    userRepo,
	}
}

func (s *contactService) AddContact(ctx context.Context, ownerID uuid.UUID, aliasName, contactEmail string) (*domain.Contact, error) {
	contactEmail = strings.ToLower(strings.TrimSpace(contactEmail))
	if aliasName == "" || contactEmail == "" {
		return nil, errors.New("alias name and contact email are required")
	}

	existingUser, err := s.userRepo.GetUserByEmail(ctx, contactEmail)
	if err != nil {
		return nil, err
	}

	newContact := &domain.Contact{
		OwnerUserID: ownerID,
		AliasName:   aliasName,
		Email:       contactEmail,
	}

	if existingUser != nil {
		if existingUser.ID == ownerID {
			return nil, errors.New("cannot add yourself as a contact")
		}
		newContact.Status = "registered"
		newContact.ContactUserID = uuid.NullUUID{UUID: existingUser.ID, Valid: true}
	} else {
		newContact.Status = "unregistered"
	}

	if err := s.contactRepo.CreateContact(ctx, newContact); err != nil {
		return nil, err
	}

	return newContact, nil
}

func (s *contactService) GetContacts(ctx context.Context, ownerID uuid.UUID) ([]*domain.Contact, error) {
	return s.contactRepo.GetContactsByOwnerID(ctx, ownerID)
}

func (s *contactService) UpdateContact(ctx context.Context, contactID, ownerID uuid.UUID, newAliasName string) error {
	if newAliasName == "" {
		return errors.New("alias name cannot be empty")
	}
	return s.contactRepo.UpdateContactAlias(ctx, contactID, ownerID, newAliasName)
}

func (s *contactService) DeleteContact(ctx context.Context, contactID, ownerID uuid.UUID) error {
	return s.contactRepo.DeleteContact(ctx, contactID, ownerID)
}

func (s *contactService) GetContactDetail(ctx context.Context, contactID, ownerID uuid.UUID) (*repository.ContactDetail, error) {
	return s.contactRepo.GetContactByID(ctx, contactID, ownerID)
}
