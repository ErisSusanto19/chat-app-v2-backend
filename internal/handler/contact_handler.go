package handler

import (
	"encoding/json"
	"net/http"

	"github.com/ErisSusanto19/chat-app-v2-backend/internal/domain"
	"github.com/ErisSusanto19/chat-app-v2-backend/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type ContactHandler struct {
	contactService service.ContactService
}

func NewContactHandler(contactService service.ContactService) *ContactHandler {
	return &ContactHandler{contactService: contactService}
}

type addContactRequest struct {
	AliasName string `json:"alias_name"`
	Email     string `json:"email"`
}

func (h *ContactHandler) AddContact(w http.ResponseWriter, r *http.Request) {
	ownerID, ok := r.Context().Value(UserContextKey).(uuid.UUID)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req addContactRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	contact, err := h.contactService.AddContact(r.Context(), ownerID, req.AliasName, req.Email)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(contact)
}

func (h *ContactHandler) GetContacts(w http.ResponseWriter, r *http.Request) {
	ownerID, ok := r.Context().Value(UserContextKey).(uuid.UUID)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	contacts, err := h.contactService.GetContacts(r.Context(), ownerID)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if contacts == nil {
		contacts = []*domain.Contact{}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(contacts)
}

type updateContactRequest struct {
	AliasName string `json:"alias_name"`
}

func (h *ContactHandler) UpdateContact(w http.ResponseWriter, r *http.Request) {
	ownerID, _ := r.Context().Value(UserContextKey).(uuid.UUID)

	contactID, err := uuid.Parse(chi.URLParam(r, "contactID"))
	if err != nil {
		http.Error(w, "Invalid contact ID", http.StatusBadRequest)
		return
	}

	var req updateContactRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	err = h.contactService.UpdateContact(r.Context(), contactID, ownerID, req.AliasName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *ContactHandler) DeleteContact(w http.ResponseWriter, r *http.Request) {
	ownerID, _ := r.Context().Value(UserContextKey).(uuid.UUID)

	contactID, err := uuid.Parse(chi.URLParam(r, "contactID"))
	if err != nil {
		http.Error(w, "Invalid contact ID", http.StatusBadRequest)
		return
	}

	err = h.contactService.DeleteContact(r.Context(), contactID, ownerID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

type contactDetailResponse struct {
	ID        uuid.UUID `json:"id"`
	AliasName string    `json:"alias_name"`
	Email     string    `json:"email"`
	Status    string    `json:"status"`
	User      *struct {
		ID    uuid.UUID `json:"id"`
		Name  string    `json:"name"`
		Image *string   `json:"image,omitempty"`
	} `json:"user,omitempty"`
}

func (h *ContactHandler) GetContactDetail(w http.ResponseWriter, r *http.Request) {
	ownerID, _ := r.Context().Value(UserContextKey).(uuid.UUID)

	contactID, err := uuid.Parse(chi.URLParam(r, "contactID"))
	if err != nil {
		http.Error(w, "Invalid contact ID", http.StatusBadRequest)
		return
	}

	detail, err := h.contactService.GetContactDetail(r.Context(), contactID, ownerID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	response := contactDetailResponse{
		ID:        detail.ID,
		AliasName: detail.AliasName,
		Email:     detail.Email,
		Status:    detail.Status,
	}
	if detail.ContactUserID.Valid {
		response.User = &struct {
			ID    uuid.UUID `json:"id"`
			Name  string    `json:"name"`
			Image *string   `json:"image,omitempty"`
		}{
			ID:    detail.ContactUserID.UUID,
			Name:  *detail.ContactName,
			Image: detail.ContactImage,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
