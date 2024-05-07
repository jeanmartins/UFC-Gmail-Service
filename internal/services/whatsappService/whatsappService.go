package whatsappservice

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	emailcontent "sigaaNotification/internal/entities/emailContent"
)

func SendEmailToWhatsapp(emails []emailcontent.EmailContent) {
	apiUrl := "https://graph.facebook.com/v18.0/251779804685576/messages"
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
		request.Header.Set("Authorization", "Bearer EAAEfy2v4qCsBOxepvgZC5eBJ6wjX159vfosgkovcHNl9nPMr3CzmWR1LsRlv4iRddZBXekd27dLi44acwR3kofbxYE5259ffLuDqmicQmfrYkXVG1K05u38PVkFmXR5ZCyYPKZAdT3nzZAkerF3uBM5TL8U30IwWpM7GHBNqoirrknJlWW8zG1CxoLz6sYzIy")
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
