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
		"subject":     "Seu código de confirmação",
		"template_id": m.TemplateID,
		"personalization": []map[string]interface{}{
			{
				"email": toEmail,
				"data": map[string]string{
					"code": code,
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

func (m *Mailer) SendPasswordResetCode(toEmail string, code string) error {
	url := "https://api.mailersend.com/v1/email"

	payload := map[string]interface{}{
		"from": map[string]string{
			"email": m.FromEmail,
			"name":  m.FromName,
		},
		"to": []map[string]string{
			{"email": toEmail},
		},
		"subject":     "Recuperação de Palavra-passe", // Assunto diferente!
		"template_id": m.TemplateID, // Vamos reaproveitar o mesmo template visual
		"personalization": []map[string]interface{}{
			{
				"email": toEmail,
				"data": map[string]string{
					"code": code,
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
		return fmt.Errorf("falha ao enviar e-mail de reset. Código HTTP: %d, Detalhes: %s", resp.StatusCode, string(bodyBytes))
	}

	return nil
}