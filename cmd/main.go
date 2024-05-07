package main

import (
	sigaaEmailService "sigaaNotification/internal/services/sigaaGmailService"
	whatsappservice "sigaaNotification/internal/services/whatsappService"
)

type EmailContent struct {
	Subject string
	Body    string
}

func main() {
	emails := sigaaEmailService.GetEmailsFromSigaa()
	whatsappservice.SendEmailToWhatsapp(emails)
}
