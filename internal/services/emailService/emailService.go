package emailService

import (
	"context"
	"encoding/base64"
	"fmt"
	"golang.org/x/oauth2"
	"time"

	"github.com/17HIERARCH70/SocialManager/internal/config"
	"github.com/17HIERARCH70/SocialManager/internal/domain/models"
	"github.com/17HIERARCH70/SocialManager/internal/services/authService"
	"github.com/jackc/pgx/v4/pgxpool"
	"golang.org/x/exp/slog"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
)

type EmailService struct {
	psql        *pgxpool.Pool
	authService *authService.AuthService
	log         *slog.Logger
	interval    time.Duration
}

func NewEmailService(psql *pgxpool.Pool, cfg *config.Config, authService *authService.AuthService, log *slog.Logger) (*EmailService, error) {
	interval, err := time.ParseDuration(cfg.Gmail.RefreshTime)
	if err != nil {
		log.Error("Failed to parse refresh interval", "error", err)
		interval = 3 * time.Minute
	}

	return &EmailService{
		psql:        psql,
		authService: authService,
		log:         log,
		interval:    interval,
	}, nil
}

func (s *EmailService) StartEmailPolling() {
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			err := s.UpdateAllEmails()
			if err != nil {
				s.log.Error("Failed to update emails", "error", err)
				return
			}
		}
	}
}

func (s *EmailService) UpdateAllEmails() error {
	users, err := s.FetchAllUsers()
	if err != nil {
		s.log.Error("Failed to fetch users", "error", err)
		return err
	}

	for _, user := range users {
		err := s.UpdateEmailsForUser(user.ID)
		if err != nil {
			s.log.Error("Failed to update emails for user", "user", user.ID, "error", err)
		}
	}
	return nil
}

func (s *EmailService) FetchAllUsers() ([]models.User, error) {
	rows, err := s.psql.Query(context.Background(), "SELECT id, google_id, email FROM users")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var user models.User
		if err := rows.Scan(&user.ID, &user.GoogleID, &user.Email); err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return users, nil
}

func (s *EmailService) UpdateEmailsForUser(userID int) error {
	// Fetch the token from the database
	token, err := s.authService.FetchGoogleTokenByUserID(userID)
	if err != nil {
		s.log.Error("Failed to fetch Google token", "user", userID, "error", err)
		return err
	}

	// Check if the token is expired
	if time.Now().After(token.Expiry) {
		// Refresh the token using the refresh token
		token, err = s.authService.RefreshToken(userID, token)
		if err != nil {
			s.log.Error("Failed to refresh token", "user", userID, "error", err)
			return err
		}
	}

	// Create the Gmail service client using the valid token
	ctx := context.Background()
	client := s.authService.OAuthConfig().Client(ctx, token)
	gmailService, err := gmail.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		s.log.Error("Failed to create Gmail service", "user", userID, "error", err)
		return err
	}

	// Fetch the user's email address from the database
	userEmail, err := s.FetchUserEmailByID(userID)
	if err != nil {
		s.log.Error("Failed to fetch user email", "user", userID, "error", err)
		return err
	}
	// Fetch emails from Gmail
	emails, err := s.FetchEmailsFromGmail(gmailService, userEmail)
	if err != nil {
		s.log.Error("Failed to fetch emails from Gmail", "user", userID, "error", err)
		return err
	}

	// Save emails to the database
	err = s.SaveEmailsToDB(token, userID, emails)
	if err != nil {
		s.log.Error("Failed to save emails to DB", "user", userID, "error", err)
		return err
	}

	return nil
}

func (s *EmailService) FetchUserEmailByID(userID int) (string, error) {
	row := s.psql.QueryRow(context.Background(), "SELECT email FROM users WHERE id=$1", userID)
	var email string
	err := row.Scan(&email)
	return email, err
}

func (s *EmailService) FetchEmailsFromGmail(service *gmail.Service, userEmail string) ([]*gmail.Message, error) {
	// Определяем время начала интервала
	startTime := time.Now().Add(-s.interval - 1*time.Minute).Unix()

	// Формируем запрос с фильтром "непрочитанные сообщения за последние interval+1 минут"
	query := fmt.Sprintf("is:unread after:%d", startTime)
	call := service.Users.Messages.List(userEmail).Q(query)
	res, err := call.Do()
	if err != nil {
		return nil, err
	}

	var emails []*gmail.Message
	for _, m := range res.Messages {
		msg, err := service.Users.Messages.Get(userEmail, m.Id).Format("full").Do()
		if err != nil {
			return nil, err
		}
		emails = append(emails, msg)
		// Изменение статуса на "прочитанное"
		modifyCall := service.Users.Messages.Modify(userEmail, m.Id, &gmail.ModifyMessageRequest{
			RemoveLabelIds: []string{"UNREAD"},
		})
		_, err = modifyCall.Do()
		if err != nil {
			return nil, err
		}
	}
	return emails, nil
}

func (s *EmailService) SaveEmailsToDB(token *oauth2.Token, userID int, messages []*gmail.Message) error {
	tx, err := s.psql.Begin(context.Background())
	if err != nil {
		return err
	}
	defer tx.Rollback(context.Background())

	for _, msg := range messages {
		body := s.extractBody(msg.Payload.Parts)
		sender := s.extractHeader(msg.Payload.Headers, "From")
		subject := s.extractHeader(msg.Payload.Headers, "Subject")
		sendedAt := s.extractSendedAt(msg.Payload.Headers)

		_, err := tx.Exec(context.Background(), `
			INSERT INTO emails (user_id, email_id, subject, body, sender, sended_at)
			VALUES ($1, $2, $3, $4, $5, $6)
			ON CONFLICT (email_id) DO NOTHING`,
			userID, msg.Id, subject, body, sender, sendedAt,
		)
		if err != nil {
			return err
		}

		for _, part := range msg.Payload.Parts {
			if part.Filename != "" {
				attachmentData := part.Body.Data
				if part.Body.AttachmentId != "" {
					attachmentData, err = s.fetchAttachment(token, msg.Id, part.Body.AttachmentId)
					if err != nil {
						s.log.Error("Failed to fetch attachment", "email_id", msg.Id, "error", err)
						return err
					}
				}

				_, err := tx.Exec(context.Background(), `
					INSERT INTO attachments (email_id, attachment_id, body, mime_type, filename)
					VALUES ((SELECT id FROM emails WHERE email_id=$1), $2, $3, $4, $5)
					ON CONFLICT (email_id, filename, mime_type) DO NOTHING`,
					msg.Id, part.Body.AttachmentId, attachmentData, part.MimeType, part.Filename,
				)
				if err != nil {
					return err
				}
			}
		}
	}

	return tx.Commit(context.Background())
}

func (s *EmailService) fetchAttachment(token *oauth2.Token, messageID, attachmentID string) (string, error) {
	client := s.authService.OAuthConfig().Client(context.Background(), token)
	gmailService, err := gmail.NewService(context.Background(), option.WithHTTPClient(client))
	if err != nil {
		s.log.Error("Failed to create Gmail service", "error", err)
		return "", err
	}

	attachment, err := gmailService.Users.Messages.Attachments.Get("me", messageID, attachmentID).Do()
	if err != nil {
		s.log.Error("Failed to fetch attachment", "error", err)
		return "", err
	}

	data, err := base64.URLEncoding.DecodeString(attachment.Data)
	if err != nil {
		s.log.Error("Failed to decode attachment data", "error", err)
		return "", err
	}

	return base64.StdEncoding.EncodeToString(data), nil
}
func (s *EmailService) extractBody(parts []*gmail.MessagePart) string {
	for _, part := range parts {
		if len(part.Parts) > 0 {
			return s.extractBody(part.Parts)
		} else {
			if part.MimeType == "text/html" {
				data, err := base64.URLEncoding.DecodeString(part.Body.Data)
				if err == nil {
					return string(data)
				}
			}
		}
	}
	return ""
}

func (s *EmailService) extractHeader(headers []*gmail.MessagePartHeader, name string) string {
	for _, header := range headers {
		if header.Name == name {
			return header.Value
		}
	}
	return ""
}

func (s *EmailService) extractSendedAt(headers []*gmail.MessagePartHeader) time.Time {
	for _, header := range headers {
		if header.Name == "Date" {
			parsedTime, err := time.Parse(time.RFC1123Z, header.Value)
			if err == nil {
				return parsedTime
			}
		}
	}
	return time.Now()
}

func (s *EmailService) GetEmailsByUserID(userID int) ([]models.Email, error) {
	rows, err := s.psql.Query(context.Background(), "SELECT id, email_id, subject, body, sender FROM emails WHERE user_id=$1", userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var emails []models.Email
	for rows.Next() {
		var email models.Email
		err := rows.Scan(&email.ID, &email.EmailID, &email.Subject, &email.Body, &email.Sender)
		if err != nil {
			return nil, err
		}
		emails = append(emails, email)
	}
	return emails, nil
}

func (s *EmailService) GetAllEmails() ([]models.Email, error) {
	rows, err := s.psql.Query(context.Background(), "SELECT id, email_id, subject, body, sender FROM emails")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var emails []models.Email
	for rows.Next() {
		var email models.Email
		err := rows.Scan(&email.ID, &email.EmailID, &email.Subject, &email.Body, &email.Sender)
		if err != nil {
			return nil, err
		}
		emails = append(emails, email)
	}
	return emails, nil
}

func (s *EmailService) GetUserIDByEmail(email string) (int, error) {
	row := s.psql.QueryRow(context.Background(), "SELECT id FROM users WHERE email=$1", email)
	var userID int
	err := row.Scan(&userID)
	return userID, err
}

func (s *EmailService) DeleteEmailByID(emailID string) error {
	tx, err := s.psql.Begin(context.Background())
	if err != nil {
		return err
	}
	defer tx.Rollback(context.Background())

	_, err = tx.Exec(context.Background(), "DELETE FROM attachments WHERE email_id=(SELECT id FROM emails WHERE email_id=$1)", emailID)
	if err != nil {
		return err
	}

	_, err = tx.Exec(context.Background(), "DELETE FROM emails WHERE email_id=$1", emailID)
	if err != nil {
		return err
	}

	return tx.Commit(context.Background())
}

func (s *EmailService) DeleteAllEmailsByUserID(userID int) error {
	tx, err := s.psql.Begin(context.Background())
	if err != nil {
		return err
	}
	defer tx.Rollback(context.Background())

	_, err = tx.Exec(context.Background(), "DELETE FROM attachments WHERE email_id IN (SELECT id FROM emails WHERE user_id=$1)", userID)
	if err != nil {
		return err
	}

	_, err = tx.Exec(context.Background(), "DELETE FROM emails WHERE user_id=$1", userID)
	if err != nil {
		return err
	}

	return tx.Commit(context.Background())
}
