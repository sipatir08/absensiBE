package utils

import (
	"fmt"
	"os"

	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

// SendEmailNotification mengirim email notifikasi ke user
func SendEmailNotification(toEmail, subject, body string) error {
	from := mail.NewEmail("Absensi App", "fathir080604@gmail.com") // Ganti dengan email pengirim
	to := mail.NewEmail("", toEmail)
	message := mail.NewSingleEmail(from, subject, to, body, body)

	client := sendgrid.NewSendClient(os.Getenv("SENDGRID_API_KEY"))
	response, err := client.Send(message)
	if err != nil {
		return err
	}
	
	if response.StatusCode >= 400 {
		return fmt.Errorf("failed to send email: %s", response.Body)
	}
	return nil
}