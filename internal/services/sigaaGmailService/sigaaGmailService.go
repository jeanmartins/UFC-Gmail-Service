package sigaaEmailService

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"os"
	emailcontent "sigaaNotification/internal/entities/emailContent"
	"sigaaNotification/internal/services/gmailService"
	"time"

	"github.com/k3a/html2text"
	"google.golang.org/api/gmail/v1"
)

func GetEmailsFromSigaa() []emailcontent.EmailContent {
	ctx := context.Background()
	gmail := gmailService.CreateGmailService(ctx)

	return getEmails(ctx, gmail)
}

func getEmails(ctx context.Context, gmailService *gmail.Service) []emailcontent.EmailContent {
	query := "is:unread from:(si3_admin_noreply@ufc.br) after:" + time.Now().Format("2006/01/02") + " before:" + time.Now().AddDate(0, 0, 1).Format("2006/01/02")
	fmt.Println("Query: ", query)
	emails, err := gmailService.Users.Messages.List("me").Q(query).Context(ctx).Do()
	if err != nil {
		fmt.Println("Error getting messages: ", err)

	}
	return formatEmailContent(ctx, emails, gmailService)
}

func formatEmailContent(ctx context.Context, emails *gmail.ListMessagesResponse, gmailService *gmail.Service) []emailcontent.EmailContent {
	var content emailcontent.EmailContent
	var emailsContent []emailcontent.EmailContent
	for _, email := range emails.Messages {
		message, err := gmailService.Users.Messages.Get("me", email.Id).Format("full").Context(ctx).Do()
		if err != nil {
			fmt.Printf("Error getting message: ", err)
			os.Exit(1)
		}
		content.Subject = getSubject(message.Payload.Headers)
		content.Body = getBody(message.Payload.Parts)
		emailsContent = append(emailsContent, content)

		response, err := gmailService.Users.Messages.Modify("me", email.Id, &gmail.ModifyMessageRequest{RemoveLabelIds: []string{"UNREAD"}}).Do()
		if err != nil {
			fmt.Println("Error marking as read: ", err)
		}
		log.Println("Marked as read: ", response)
	}
	return emailsContent
}

func getSubject(headers []*gmail.MessagePartHeader) string {
	for _, header := range headers {
		if header.Name == "Subject" {
			return header.Value
		}
	}
	return ""
}

func getBody(parts []*gmail.MessagePart) string {
	for _, part := range parts {
		if part.MimeType == "text/html" || part.MimeType == "text/plain" {
			data, _ := base64.URLEncoding.DecodeString(part.Body.Data)
			html := string(data)
			return html2text.HTML2Text(html)
		}
	}
	return ""
}
