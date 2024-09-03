package client

import (
	"encoding/json"
	"github.com/M0hammadUsman/letschat/internal/domain"
	"io"
	"log"
	"net/http"
	"strconv"
	"sync"
)

var (
	once   sync.Once
	client *Client
)

type Client struct {
	AuthToken string // if zero requires login
	krm       *keyringManager
}

func Init() error {
	var c Client
	var err error
	once.Do(func() {
		c.krm, err = newKeyringManager()
		// ignoring the error, we'll determine if the item is not found using zero value of Client.AuthToken
		c.AuthToken, _ = c.krm.getAuthTokenFromKeyring()
	})
	if err != nil {
		return err
	}
	client = &c
	return nil
}

func Get() *Client {
	return client
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
		return nil, getMostNestedError(err), 503
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
		return nil, getMostNestedError(err), 503
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
	log.Println(response)
	return &response.User, nil, resp.StatusCode
}
