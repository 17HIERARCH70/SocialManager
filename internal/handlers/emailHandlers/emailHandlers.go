package emailHandlers

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"golang.org/x/exp/slog"
	"net/http"
	"strconv"

	"github.com/17HIERARCH70/SocialManager/internal/services/emailService"
)

type EmailHandler struct {
	emailService *emailService.EmailService
	log          *slog.Logger
}

func NewEmailHandler(emailService *emailService.EmailService, log *slog.Logger) *EmailHandler {
	return &EmailHandler{
		emailService: emailService,
		log:          log,
	}
}

func (h *EmailHandler) UpdateEmailsHandler(w http.ResponseWriter, r *http.Request) {
	err := h.emailService.UpdateAllEmails()
	if err != nil {
		h.log.Error("Failed to update all emails", "error", err)
		http.Error(w, "Failed to update emails", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Emails updated successfully"))
}

func (h *EmailHandler) GetEmailsByUserIDHandler(w http.ResponseWriter, r *http.Request) {
	userID, err := strconv.Atoi(mux.Vars(r)["user_id"])
	if err != nil {
		h.log.Error("Invalid user ID", "error", err)
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	emails, err := h.emailService.GetEmailsByUserID(userID)
	if err != nil {
		h.log.Error("Failed to get emails", "error", err)
		http.Error(w, "Failed to get emails", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(emails)
	if err != nil {
		h.log.Error("Failed to encode response", "error", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

func (h *EmailHandler) GetAllEmailsHandler(w http.ResponseWriter, r *http.Request) {
	emails, err := h.emailService.GetAllEmails()
	if err != nil {
		h.log.Error("Failed to get all emails", "error", err)
		http.Error(w, "Failed to get all emails", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(emails)
	if err != nil {
		h.log.Error("Failed to encode response", "error", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

func (h *EmailHandler) GetUserIDByEmailHandler(w http.ResponseWriter, r *http.Request) {
	email := r.URL.Query().Get("email")
	if email == "" {
		http.Error(w, "Email is required", http.StatusBadRequest)
		return
	}

	userID, err := h.emailService.GetUserIDByEmail(email)
	if err != nil {
		h.log.Error("Failed to get user ID", "error", err)
		http.Error(w, "Failed to get user ID", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(map[string]int{"user_id": userID})
	if err != nil {
		h.log.Error("Failed to encode response", "error", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

func (h *EmailHandler) DeleteEmailByIDHandler(w http.ResponseWriter, r *http.Request) {
	emailID := mux.Vars(r)["email_id"]
	if emailID == "" {
		http.Error(w, "Email ID is required", http.StatusBadRequest)
		return
	}

	err := h.emailService.DeleteEmailByID(emailID)
	if err != nil {
		h.log.Error("Failed to delete email", "error", err)
		http.Error(w, "Failed to delete email", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Email deleted successfully"))
}

func (h *EmailHandler) DeleteAllEmailsByUserIDHandler(w http.ResponseWriter, r *http.Request) {
	userID, err := strconv.Atoi(mux.Vars(r)["user_id"])
	if err != nil {
		h.log.Error("Invalid user ID", "error", err)
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	err = h.emailService.DeleteAllEmailsByUserID(userID)
	if err != nil {
		h.log.Error("Failed to delete all emails", "error", err)
		http.Error(w, "Failed to delete all emails", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("All emails deleted successfully"))
}
