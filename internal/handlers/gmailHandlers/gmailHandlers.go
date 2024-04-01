package gmailHandlers

import (
	"github.com/17HIERARCH70/SocialManager/internal/services"
	"golang.org/x/exp/slog"
	"net/http"
)

type GmailHandler struct {
	GmailService services.GmailServiceMethods
	Log          *slog.Logger
}

func (h *GmailHandler) FetchEmailsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	err := h.GmailService.FetchAndStoreEmails(ctx)
	if err != nil {
		// Log the error
		h.Log.Error("Failed to fetch and store emails", "error", err)

		// Inform the client of the failure
		http.Error(w, "Failed to fetch emails", http.StatusInternalServerError)
		return
	}

	// If successful, send back a success response
	h.Log.Info("Emails fetched and stored successfully")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Emails fetched and stored successfully"))

}
