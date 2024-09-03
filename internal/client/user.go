package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/M0hammadUsman/letschat/internal/domain"
	"io"
	"log"
	"net/http"
)

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
		return getMostNestedError(err)
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

func (c *Client) Login(u domain.UserAuth) error {
	b, err := json.Marshal(u)
	if err != nil {
		log.Println(err.Error())
		return err
	}
	res, err := http.DefaultClient.Post(authenticate, "application/json", bytes.NewBuffer(b))
	if err != nil {
		log.Println(err.Error())
		return getMostNestedError(err)
	}
	defer res.Body.Close()
	readBody, err := io.ReadAll(res.Body)
	if err != nil {
		log.Println(err.Error())
		return err
	}
	if res.StatusCode == http.StatusOK {
		var token struct {
			Token string `json:"token"`
		}
		if err = json.Unmarshal(readBody, &token); err != nil {
			log.Println(err.Error())
			return err
		}
		c.AuthToken = token.Token
	} else {
		var ev struct {
			Errors *domain.UserAuth `json:"errors"`
		}
		if err = json.Unmarshal(readBody, &ev); err != nil {
			log.Println(err.Error())
			return err
		}
		if ev.Errors.Email == ErrNonActiveUser.Error() {
			return ErrNonActiveUser
		} else {
			return ErrUnauthorized
		}
	}
	// putting auth token in keyring
	if err = c.krm.setAuthTokenInKeyring(u.Email, c.AuthToken); err != nil {
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
		return getMostNestedError(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		log.Println(res.Status)
		resBody, _ := io.ReadAll(res.Body)
		_ = json.Unmarshal(resBody, &token)
		return ErrExpiredOTP
	}
	return nil
}
