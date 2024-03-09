package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/k3a/html2text"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
)

type EmailContent struct {
	Subject string
	Body    string
}

func main() {

	var emailsContent []EmailContent
	ctx := context.Background()
	gmailService := createGmailService(ctx)
	emails := getEmails(ctx, gmailService, emailsContent)

	for _, email := range emails {
		fmt.Println("Subject: ", email.Subject)
		fmt.Println("Body: ", email.Body)
		fmt.Println("---------------------------------------------------")
	}

}

func getEmails(ctx context.Context, gmailService *gmail.Service, emailsContent []EmailContent) []EmailContent {
	query := "is:unread from:(si3_admin_noreply@ufc.br) after:2024/3/8 before:2024/3/11"
	emails, err := gmailService.Users.Messages.List("me").Q(query).Context(ctx).Do()
	if err != nil {
		fmt.Println("Error getting messages: ", err)

	}
	return getEmailsContent(ctx, emails, gmailService)
}

func getEmailsContent(ctx context.Context, emails *gmail.ListMessagesResponse, gmailService *gmail.Service) []EmailContent {
	var content EmailContent
	var emailsContent []EmailContent
	for _, email := range emails.Messages {
		message, err := gmailService.Users.Messages.Get("me", email.Id).Format("full").Context(ctx).Do()
		if err != nil {
			log.Fatalf("Error getting message: ", err)
		}
		content.Subject = getSubject(message.Payload.Headers)
		content.Body = getBody(message.Payload.Parts)
		emailsContent = append(emailsContent, content)
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
func createGmailService(ctx context.Context) *gmail.Service {

	credentials, err := os.ReadFile("credentials.json")
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}
	config, err := google.ConfigFromJSON(credentials, gmail.GmailReadonlyScope)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config %v", err)
	}
	client := getClient(config)

	gmailService, err := gmail.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		fmt.Println("Error creating Gmail service: ", err)
	}
	return gmailService
}

func getClient(config *oauth2.Config) *http.Client {
	tokFile := "token.json"
	tok, err := tokenFromFile(tokFile)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(tokFile, tok)
	}
	return config.Client(context.Background(), tok)
}

func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		log.Fatalf("Unable to read authorization code: %v", err)
	}

	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web: %v", err)
	}
	return tok
}

// Retrieves a token from a local file.
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

// Saves a token to a file path.
func saveToken(path string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}
