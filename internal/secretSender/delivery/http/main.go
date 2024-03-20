package http

import (
	"fmt"
	"net/http"
)

type SecretSender struct {
	client *http.Client
}

func NewSecretSender(c *http.Client) SecretSender {
	return SecretSender{
		client: c,
	}
}

func (s *SecretSender) Send() {
	req, err := http.NewRequest("GET", "https://api.example.com/data", nil)
	if err != nil {
		fmt.Println("Error creating request:", err)
		return
	}
}
