package sender

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"log/slog"
	"mime"
	"net/smtp"
	"strings"
	"time"

	"github.com/appraisal-crm/notification-service/internal/domain"
)

type SMTPSender struct {
	host     string
	port     int
	user     string
	password string
	from     string
	enabled  bool
}

func NewSMTPSender(host string, port int, user, password, from string, enabled bool) *SMTPSender {
	return &SMTPSender{
		host:     host,
		port:     port,
		user:     user,
		password: password,
		from:     from,
		enabled:  enabled,
	}
}

func (s *SMTPSender) Send(ctx context.Context, n *domain.Notification) error {
	if !s.enabled {
		slog.InfoContext(ctx, "SMTP is disabled — skipping email delivery",
			"notification_id", n.ID,
			"recipient", n.RecipientAddress,
			"event_type", n.EventType,
		)
		return nil
	}

	if n.RecipientAddress == nil || strings.TrimSpace(*n.RecipientAddress) == "" {
		return errors.New("recipient address is empty")
	}
	to := strings.TrimSpace(*n.RecipientAddress)

	subject := "Уведомление"
	if n.Subject != nil && *n.Subject != "" {
		subject = *n.Subject
	}

	encodedSubject := mime.BEncoding.Encode("UTF-8", subject)

	headers := make(map[string]string)
	headers["From"] = s.from
	headers["To"] = to
	headers["Subject"] = encodedSubject
	headers["MIME-Version"] = "1.0"
	headers["Content-Type"] = "text/html; charset=UTF-8"
	headers["Date"] = time.Now().UTC().Format(time.RFC1123Z)

	var msg strings.Builder
	for k, v := range headers {
		msg.WriteString(fmt.Sprintf("%s: %s\r\n", k, v))
	}
	msg.WriteString("\r\n")
	msg.WriteString(n.Body)

	addr := fmt.Sprintf("%s:%d", s.host, s.port)

	slog.InfoContext(ctx, "sending email via SMTP",
		"to", to,
		"subject", subject,
		"smtp_addr", addr,
	)

	// Direct send for SSL (port 465) vs standard SMTP (port 25, 587, 1025)
	if s.port == 465 {
		return s.sendSSL(addr, to, []byte(msg.String()))
	}

	return s.sendStandard(addr, to, []byte(msg.String()))
}

func (s *SMTPSender) sendStandard(addr, to string, body []byte) error {
	var auth smtp.Auth
	if s.user != "" && s.password != "" {
		auth = smtp.PlainAuth("", s.user, s.password, s.host)
	}

	return smtp.SendMail(addr, auth, s.from, []string{to}, body)
}

func (s *SMTPSender) sendSSL(addr, to string, body []byte) error {
	tlsConfig := &tls.Config{
		ServerName: s.host,
	}

	conn, err := tls.Dial("tcp", addr, tlsConfig)
	if err != nil {
		return fmt.Errorf("tls dial failed: %w", err)
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, s.host)
	if err != nil {
		return fmt.Errorf("smtp client creation failed: %w", err)
	}
	defer client.Close()

	if s.user != "" && s.password != "" {
		auth := smtp.PlainAuth("", s.user, s.password, s.host)
		if err := client.Auth(auth); err != nil {
			return fmt.Errorf("smtp auth failed: %w", err)
		}
	}

	if err := client.Mail(s.from); err != nil {
		return fmt.Errorf("mail from command failed: %w", err)
	}

	if err := client.Rcpt(to); err != nil {
		return fmt.Errorf("rcpt to command failed: %w", err)
	}

	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("data command failed: %w", err)
	}

	if _, err := w.Write(body); err != nil {
		return fmt.Errorf("write body failed: %w", err)
	}

	if err := w.Close(); err != nil {
		return fmt.Errorf("close writer failed: %w", err)
	}

	return client.Quit()
}

// Compile-time check that SMTPSender implements Sender
var _ Sender = (*SMTPSender)(nil)
