package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/M0hammadUsman/letschat/internal/domain"
	"io"
	"log"
	"net/http"
	"strings"
)

const serverDownErr = "No connection could be made because the target machine actively refused it."

// Register will register the user & populate the *domain.UserRegister with validation errors
// in case of http.StatusUnprocessableEntity
func (*Client) Register(u *domain.UserRegister) error {
	body, err := json.Marshal(u)
	if err != nil {
		log.Println(err.Error())
		return err
	}
	resp, err := http.DefaultClient.Post(registerUser, "application/json", bytes.NewBuffer(body))
	if err != nil {
		log.Println(err.Error())
		if strings.Contains(err.Error(), serverDownErr) {
			return errors.New(serverDownErr)
		}
		return err
	}
	defer resp.Body.Close()
	switch resp.StatusCode {
	case http.StatusUnprocessableEntity:
		var ev struct {
			Errors *domain.UserRegister `json:"errors"`
		}
		ev.Errors = u
		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Println(err.Error())
			return err
		}
		if err = json.Unmarshal(respBody, &ev); err != nil {
			log.Println(err.Error())
			return err
		}
		return ErrServerValidation
	case http.StatusInternalServerError:
		return errors.New("the server is overwhelmed")
	}
	return nil
}

func (c *Client) Login(email, password string) error {
	data := domain.UserAuth{
		Email:    email,
		Password: password,
	}
	b, err := json.Marshal(data)
	if err != nil {
		log.Println(err.Error())
		return err
	}
	res, err := http.DefaultClient.Post(authenticate, "application/json", bytes.NewBuffer(b))
	if err != nil {
		log.Println(err.Error())
		return err
	}
	defer res.Body.Close()
	var resBody []byte
	if _, err = res.Body.Read(resBody); err != nil {
		log.Println(err.Error())
		return err
	}
	if err = json.Unmarshal(resBody, &c.AuthToken); err != nil {
		log.Println(err.Error())
		return err
	}
	// putting auth token in keyring
	if err = c.krm.setAuthTokenInKeyring(c.AuthToken); err != nil {
		log.Println(err.Error())
		return err
	}
	return nil
}

func (c *Client) ActivateUser(otp string) error {
	var token struct {
		OTP string `json:"otp"`
	}
	token.OTP = otp
	jsonBytes, err := json.Marshal(token)
	if err != nil {
		log.Println(err.Error())
		return err
	}
	res, err := http.Post(activateUser, "application/json", bytes.NewBuffer(jsonBytes))
	if err != nil {
		log.Println(err.Error())
		if strings.Contains(err.Error(), serverDownErr) {
			return errors.New(serverDownErr)
		}
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		log.Println(res.Status)
		log.Printf("%+v", res)
		resBody, _ := io.ReadAll(res.Body)
		_ = json.Unmarshal(resBody, &token)
		log.Printf("%+v", token)
		return ErrExpiredOTP
	}
	return nil
}
