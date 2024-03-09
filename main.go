package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
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
	apiUrl := "https://graph.facebook.com/v18.0/251779804685576/messages"
	ctx := context.Background()
	gmailService := createGmailService(ctx)
	emails := getEmails(ctx, gmailService, emailsContent)

	for _, email := range emails {
		fmt.Println("Subject: ", email.Subject)
		fmt.Println("Body: ", email.Body)
		fmt.Println("---------------------------------------------------")
		message := map[string]interface{}{
			"messaging_product": "whatsapp",
			"recipient_type":    "individual",
			"to":                "5585986566632",
			"type":              "text",
			"text": map[string]interface{}{
				"body": email.Body,
			},
		}
		jsonMessage, err := json.Marshal(message)
		if err != nil {
			fmt.Println("Error marshalling message: ", err)
		}
		request, _ := http.NewRequest("POST", apiUrl, bytes.NewBuffer(jsonMessage))
		request.Header.Set("Authorization", "Bearer EAAEfy2v4qCsBOyjhaDgZCOz1fLoFBxJbPLPZAqYhcfQqDNOO84WxizGYV8CZBzQeVW3x5LYeDRPnIlvW9t4bUcJ6FyttFGIOOxnQxZBxjZCAv2iMWd56nTevOySu0mlPMpZCTGwoW88zYZAziEj1RT19oLqVQwaGvSpAnbwKebTBCWHEzyaZC4bPaYj6uxzSKb2UDakBKEYjnBPHJ030VJcZD")
		request.Header.Set("Content-Type", "application/json")
		client := &http.Client{}
		response, error := client.Do(request)
		if error != nil {
			fmt.Println("Error sending message: ", error)
		}
		responseBody, error := io.ReadAll(response.Body)
		if error != nil {
			fmt.Println(error)
		}
		formattedData := formatJSON(responseBody)
		fmt.Println("Status: ", response.Status)
		fmt.Println("Response body: ", formattedData)

		// clean up memory after execution
		defer response.Body.Close()
	}
}
func formatJSON(data []byte) string {
	var out bytes.Buffer
	err := json.Indent(&out, data, "", " ")

	if err != nil {
		fmt.Println(err)
	}

	d := out.Bytes()
	return string(d)
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
