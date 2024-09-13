package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
)

func (c *Client) ResendOtp(email string) error {
	body := struct {
		Email string `json:"email"`
	}{Email: email}
	jsonBytes, err := json.Marshal(body)
	if err != nil {
		slog.Error(err.Error())
		return err
	}
	resp, err := http.DefaultClient.Post(generateOTP, "application/json", bytes.NewBuffer(jsonBytes))
	if err != nil {
		slog.Error(err.Error())
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusAccepted {
		return errors.New(http.StatusText(resp.StatusCode))
	}
	return nil
}
