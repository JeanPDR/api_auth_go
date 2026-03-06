package mailer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

type Mailer struct {
	APIKey     string
	FromEmail  string
	FromName   string
	TemplateID string
}

func NewMailer() *Mailer {
	return &Mailer{
		APIKey:     os.Getenv("MAILERSEND_API_KEY"),
		FromEmail:  os.Getenv("MAILERSEND_FROM_EMAIL"),
		FromName:   os.Getenv("MAILERSEND_FROM_NAME"),
		TemplateID: os.Getenv("MAILERSEND_TEMPLATE_ID"),
	}
}

func (m *Mailer) SendConfirmationCode(toEmail string, code string) error {
	url := "https://api.mailersend.com/v1/email"

	payload := map[string]interface{}{
		"from": map[string]string{
			"email": m.FromEmail,
			"name":  m.FromName,
		},
		"to": []map[string]string{
			{"email": toEmail},
		},
		"subject": "Seu código de confirmação", // 🚨 ADICIONE ESTA LINHA AQUI
		"template_id": m.TemplateID,
		"variables": []map[string]interface{}{
			{
				"email": toEmail,
				"substitutions": []map[string]string{
					{
						"var":   "code",
						"value": code,
					},
				},
			},
		},
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Requested-With", "XMLHttpRequest")
	req.Header.Set("Authorization", "Bearer "+m.APIKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("falha ao enviar e-mail. Código HTTP: %d, Detalhes: %s", resp.StatusCode, string(bodyBytes))
	}

	return nil
}