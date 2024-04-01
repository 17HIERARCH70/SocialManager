package authHandlers

import (
	"encoding/json"
	"github.com/17HIERARCH70/SocialManager/internal/domain/models"
	"github.com/17HIERARCH70/SocialManager/internal/lib/jwt"
	"github.com/17HIERARCH70/SocialManager/internal/services"
	"golang.org/x/exp/slog"
	"net/http"
	"time"
)

type AuthHandler struct {
	UserService services.UserServiceMethods
	Log         *slog.Logger
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var creds models.Credentials // Assume this struct exists in your domain/models package
	err := json.NewDecoder(r.Body).Decode(&creds)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	user, err := h.UserService.AuthenticateUser(creds.Email, creds.Password)
	if err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	token, err := jwt.NewToken(*user, time.Hour*24)
	if err != nil {
		h.Log.Error("Failed to generate token", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"token": token,
	})
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var newUser models.User
	err := json.NewDecoder(r.Body).Decode(&newUser)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	var id int
	id, err = h.UserService.CreateUser(&newUser)
	if err != nil {
		h.Log.Error("Failed to create user", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	h.Log.Info("User created", "ID", id, "user", newUser.Email)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
}
