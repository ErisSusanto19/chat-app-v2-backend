package handler

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/ErisSusanto19/chat-app-v2-backend/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type AuthHandler struct {
	authService service.AuthService
}

func NewAuthHandler(authService service.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

func (h *AuthHandler) RegisterRoutes(router *chi.Mux) {
	router.Post("/register", h.Register)
	router.Post("/login", h.Login)
}

type registerRequest struct {
	Name     string `json:"name" validate:"required,min=2"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := BindAndValidate(r, &req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	user, err := h.authService.Register(r.Context(), req.Name, req.Email, req.Password)
	// if err != nil {
	// 	log.Printf("DEBUG: Error from Register service in handler: %v", err)
	// 	if errors.Is(err, errors.New("user with this email already exists")) {
	// 		http.Error(w, err.Error(), http.StatusConflict)
	// 	} else if errors.Is(err, errors.New("name and email cannot be empty")) || errors.Is(err, errors.New("password must be at least 8 characters long")) {
	// 		http.Error(w, err.Error(), http.StatusBadRequest)
	// 	} else {
	// 		http.Error(w, "Internal server error", http.StatusInternalServerError)
	// 	}
	// 	return
	// }

	if err != nil {
		if strings.Contains(err.Error(), "user with this email already exists") {
			http.Error(w, err.Error(), http.StatusConflict)
		} else {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginResponse struct {
	Token string `json:"token"`
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	token, err := h.authService.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			http.Error(w, err.Error(), http.StatusUnauthorized)
		} else {
			log.Printf("DEBUG: Unexpected error during login: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(loginResponse{Token: token})
}

func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(UserContextKey).(uuid.UUID)
	if !ok {
		http.Error(w, "Could not retrieve user from context", http.StatusInternalServerError)
		return
	}

	user, err := h.authService.GetProfile(r.Context(), userID)
	if err != nil {
		if errors.Is(err, errors.New("user not found")) {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(user)
}

type updateProfileRequest struct {
	Name          string  `json:"name"`
	PhoneNumber   *string `json:"phone_number"`
	ImagePublicID *string `json:"image_public_id"`
}

func (h *AuthHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	userID, _ := r.Context().Value(UserContextKey).(uuid.UUID)

	var req updateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	updatedUser, err := h.authService.UpdateProfile(r.Context(), userID, req.Name, req.PhoneNumber, req.ImagePublicID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(updatedUser)
}

type changePasswordRequest struct {
	OldPassword string `json:"old_password" validate:"required"`
	NewPassword string `json:"new_password" validate:"required,min=8"`
}

func (h *AuthHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	userID, _ := r.Context().Value(UserContextKey).(uuid.UUID)

	var req changePasswordRequest
	if err := BindAndValidate(r, &req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err := h.authService.ChangePassword(r.Context(), userID, req.OldPassword, req.NewPassword)
	// if err != nil {
	// 	if err.Error() == "incorrect old password" {
	// 		http.Error(w, err.Error(), http.StatusUnauthorized)
	// 	} else {
	// 		http.Error(w, err.Error(), http.StatusBadRequest)
	// 	}
	// 	return
	// }

	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			http.Error(w, err.Error(), http.StatusUnauthorized)
		} else {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
