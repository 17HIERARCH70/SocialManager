package services

import (
	"context"
	"encoding/base64"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v4/pgxpool"
	"golang.org/x/exp/slog"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
	"net"
	"net/http"
	"os"
	"time"
)

// GmailServiceMethods outlines methods to interact with Gmail
type GmailServiceMethods interface {
	FetchAndStoreEmails(ctx context.Context) error
}

type gmailService struct {
	Log     *slog.Logger
	service *gmail.Service
	db      *pgxpool.Pool // Assuming you have a database connection pool available
}

// NewGmailService initializes a new instance of Gmail service with the given OAuth2 client and database connection.
// A logger is also passed for logging purposes.
func NewGmailService(ctx context.Context, secretPath string, db *pgxpool.Pool, logger *slog.Logger) GmailServiceMethods {
	b, err := os.ReadFile(secretPath)
	if err != nil {
		logger.Error(fmt.Sprintf("Unable to read client secret file: %v", err))
		return nil
	}
	config, err := google.ConfigFromJSON(b, gmail.GmailReadonlyScope)
	if err != nil {
		logger.Error(fmt.Sprintf("Unable to parse client secret file to config: %v", err))
		return nil
	}
	token := getTokenFromWeb(config)
	client := config.Client(ctx, token)
	srv, err := gmail.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		logger.Error(fmt.Sprintf("Unable to retrieve Gmail client: %v", err))
		return nil
	}

	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	logger.Info("Please authorize app by visiting this URL: ", authURL)

	return &gmailService{
		Log:     logger,
		service: srv,
		db:      db,
	}
}

// Получает токен из файла, если он существует, или запускает процесс авторизации.
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	ch := make(chan string)
	defer close(ch)
	r := mux.NewRouter()
	r.HandleFunc("/api/gmail/token", func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		_, _ = fmt.Fprintf(w, "Succsessfully retrieved authorization code. You can close this window now.")
		ch <- code
	})
	listener, err := net.Listen("tcp", "localhost:8080")
	if err != nil {
		slog.Error("Unable to start the local server for automatic authorization code retrieval: %v", err)
		return nil
	}

	go func() {
		_ = http.Serve(listener, r)
	}()

	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Open the following link in your browser: %v\n", authURL)

	authCode := <-ch

	token, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		slog.Error("Unable to exchange authorization code for token: %v", err)
		return nil
	}

	_ = listener.Close()
	return token
}

func (s *gmailService) FetchAndStoreEmails(ctx context.Context) error {
	user := "me" // Используйте "me" для обозначения аутентифицированного пользователя.
	req := s.service.Users.Messages.List(user).Q("newer_than:1h")
	r, err := req.Do()
	if err != nil {
		s.Log.Error("Unable to retrieve messages", "error", err)
		return err
	}

	for _, m := range r.Messages {
		msg, err := s.service.Users.Messages.Get(user, m.Id).Do()
		if err != nil {
			s.Log.Error("Unable to retrieve message", "messageID", m.Id, "error", err)
			continue
		}

		err = storeEmail(s.db, msg)
		if err != nil {
			s.Log.Error("Failed to store email", "messageID", m.Id, "error", err)
			continue
		}
	}

	return nil
}

func storeEmail(db *pgxpool.Pool, msg *gmail.Message) error {
	var sender, subject, body, receivedAt string

	for _, header := range msg.Payload.Headers {
		switch header.Name {
		case "From":
			sender = header.Value
		case "Subject":
			subject = header.Value
		case "Date":
			receivedAt = header.Value
		}
	}

	// Attempt to parse and decode the email body
	if msg.Payload.MimeType == "text/plain" && len(msg.Payload.Parts) == 0 {
		// Decode base64 URL-encoded body
		decodedBody, err := base64.URLEncoding.DecodeString(msg.Payload.Body.Data)
		if err != nil {
			slog.Error("failed to decode email body: %w", err)
		}
		body = string(decodedBody)
	} else if msg.Payload.MimeType == "multipart/alternative" && len(msg.Payload.Parts) > 0 {
		part := msg.Payload.Parts[0]
		decodedPart, err := base64.URLEncoding.DecodeString(part.Body.Data)
		if err != nil {
			slog.Error("failed to decode multipart email body: %w", err)
		}
		body = string(decodedPart)
		return err
	}

	var parsedTime time.Time
	var err error
	formats := []string{time.RFC1123Z, time.RFC1123, "Mon, 2 Jan 2006 15:04:05 -0700", "Mon, 2 Jan 2006 15:04:05 MST"}
	for _, format := range formats {
		parsedTime, err = time.Parse(format, receivedAt)
		if err == nil {
			break
		}
	}
	if err != nil {
		return fmt.Errorf("failed to parse date: %w", err)
	}

	// Insert into the database
	sql := `INSERT INTO emails (sender, subject, body, received_at) VALUES ($1, $2, $3, $4)`
	_, err = db.Exec(context.Background(), sql, sender, subject, body, parsedTime)
	if err != nil {
		slog.Error("failed to insert email into database: %w", err)
	}

	return nil
}
