package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/M0hammadUsman/letschat/internal/domain"
	"io"
	"log"
	"net/http"
	"strconv"
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

type PagedUserResponse struct {
	Metadata domain.Metadata `json:"metadata"`
	Users    []domain.User   `json:"users"`
}

func (c *Client) SearchUser(param string, page int) (*PagedUserResponse, error, int) {
	r, err := http.NewRequest(http.MethodGet, searchUser, nil)
	if err != nil {
		log.Println(err.Error())
		return nil, ErrApplication, 0
	}
	r.Header.Set("Authorization", "Bearer "+c.AuthToken)
	v := r.URL.Query()
	v.Set("param", param)
	v.Set("page", strconv.Itoa(page))
	r.URL.RawQuery = v.Encode()
	resp, err := http.DefaultClient.Do(r)
	if err != nil {
		log.Println(err.Error())
		return nil, getMostNestedError(err), http.StatusServiceUnavailable
	}
	defer resp.Body.Close()
	readBody, _ := io.ReadAll(resp.Body)
	var pur PagedUserResponse
	if err = json.Unmarshal(readBody, &pur); err != nil {
		log.Println(err.Error())
		return nil, ErrApplication, 0
	}
	return &pur, nil, resp.StatusCode
}

func (c *Client) GetCurrentActiveUser() (*domain.User, error, int) {
	r, err := http.NewRequest(http.MethodGet, getCurrentActiveUser, nil)
	if err != nil {
		log.Println(err.Error())
		return nil, ErrApplication, 0
	}
	r.Header.Set("Authorization", "Bearer "+c.AuthToken)
	resp, err := http.DefaultClient.Do(r)
	if err != nil {
		log.Println(err.Error())
		return nil, getMostNestedError(err), http.StatusServiceUnavailable
	}
	defer resp.Body.Close()
	readBody, _ := io.ReadAll(resp.Body)
	var response struct {
		User domain.User `json:"user"`
	}
	if err = json.Unmarshal(readBody, &response); err != nil {
		log.Println(err.Error())
		return nil, ErrApplication, 0
	}
	return &response.User, nil, resp.StatusCode
}
