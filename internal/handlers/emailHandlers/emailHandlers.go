package emailHandlers

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"golang.org/x/exp/slog"
	"net/http"
	"strconv"

	"github.com/17HIERARCH70/SocialManager/internal/services/emailService"
)

// EmailHandler handles email-related requests
type EmailHandler struct {
	emailService *emailService.EmailService
	log          *slog.Logger
}

// NewEmailHandler creates a new EmailHandler
func NewEmailHandler(emailService *emailService.EmailService, log *slog.Logger) *EmailHandler {
	return &EmailHandler{
		emailService: emailService,
		log:          log,
	}
}

// UpdateEmailsHandler updates all emails
// @Summary Update Emails
// @Description Update all emails
// @Tags emails
// @Produce plain
// @Success 200 {string} string "Emails updated successfully"
// @Router /emails/update [put]
func (h *EmailHandler) UpdateEmailsHandler(w http.ResponseWriter, r *http.Request) {
	err := h.emailService.UpdateAllEmails()
	if err != nil {
		h.log.Error("Failed to update all emails", "error", err)
		http.Error(w, "Failed to update emails", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	_, err = w.Write([]byte("Emails updated successfully"))
	if err != nil {
		return
	}
}

// GetEmailsByUserIDHandler retrieves emails by user ID
// @Summary Get Emails by User ID
// @Description Retrieve emails by user ID
// @Tags emails
// @Produce json
// @Param user_id path int true "User ID"
// @Success 200 {array} models.Email
// @Router /emails/user/{user_id} [get]
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

// GetAllEmailsHandler retrieves all emails
// @Summary Get All Emails
// @Description Retrieve all emails
// @Tags emails
// @Produce json
// @Success 200 {array} models.Email
// @Router /emails [get]
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

// GetUserIDByEmailHandler retrieves user ID by email
// @Summary Get User ID by Email
// @Description Retrieve user ID by email
// @Tags emails
// @Produce json
// @Param email query string true "Email"
// @Success 200 {object} map[string]int "user_id"
// @Router /emails/user [get]
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

// DeleteEmailByIDHandler deletes an email by ID
// @Summary Delete Email by ID
// @Description Delete an email by its ID
// @Tags emails
// @Produce plain
// @Param email_id path string true "Email ID"
// @Success 200 {string} string "Email deleted successfully"
// @Router /emails/{email_id} [delete]
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
	_, err = w.Write([]byte("Email deleted successfully"))
	if err != nil {
		return
	}
}

// DeleteAllEmailsByUserIDHandler deletes all emails by user ID
// @Summary Delete All Emails by User ID
// @Description Delete all emails by user ID
// @Tags emails
// @Produce plain
// @Param user_id path int true "User ID"
// @Success 200 {string} string "All emails deleted successfully"
// @Router /emails/user/{user_id} [delete]
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
	_, err = w.Write([]byte("All emails deleted successfully"))
	if err != nil {
		h.log.Error("Failed to write response", "error", err)
		http.Error(w, "Failed to write response", http.StatusInternalServerError)
		return
	}
}
