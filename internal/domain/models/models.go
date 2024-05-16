package models

import (
	"time"
)

// User represents a user in the system.
type User struct {
	ID         int    `json:"id"`
	GoogleID   string `json:"google_id"`
	TelegramID string `json:"telegram_id"`
	Email      string `json:"email"`
}

// Email represents an email fetched from Gmail.
type Email struct {
	ID         int          `json:"id"`
	EmailID    string       `json:"email_id"`
	Subject    string       `json:"subject"`
	Body       string       `json:"body"`
	Sender     string       `json:"sender"`
	Attachment []Attachment `json:"attachment"`
	CreatedAt  time.Time    `json:"created_at"`
}

// Attachment represents an attachment in an email.
type Attachment struct {
	ID       int    `json:"id"`
	EmailID  int    `json:"email_id"`
	Body     string `json:"body"`
	File     string `json:"file"`
	MimeType string `json:"mime_type"`
	Filename string `json:"filename"`
}

//// HTML will return the value that can be used to put in a HTML attribute.
//func (a Attachment) HTML() string {
//	data, _ := base64.URLEncoding.DecodeString(a.Body)
//	html := base64.StdEncoding.EncodeToString(data)
//	return fmt.Sprintf("data:%s;base64,%s", a.MimeType, html)
//}
//
//// Placeholder will return the placeholder that is used in an email to identify
//// the attachment.
//func (a Attachment) Placeholder() string {
//	return fmt.Sprintf("cid:%s", a.Filename)
//}
//
//// HTML the decoded email body as HTML contents and an error if decoding failed.
//func (e Email) HTML() (html string, err error) {
//	data, err := base64.URLEncoding.DecodeString(e.Body)
//
//	if err == nil {
//		html = string(data)
//
//		if len(e.Attachments) > 0 {
//			html = replaceAttachments(html, e.Attachments)
//		}
//	}
//
//	return html, err
//}
//
//func replaceAttachments(html string, attachments []Attachment) string {
//	for _, attachment := range attachments {
//		html = strings.Replace(
//			html,
//			attachment.Placeholder(),
//			attachment.HTML(),
//			-1,
//		)
//	}
//
//	return html
//}
