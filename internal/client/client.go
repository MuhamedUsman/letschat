package client

import (
	"errors"
	"fmt"
	"github.com/M0hammadUsman/letschat/internal/client/repository"
	"github.com/M0hammadUsman/letschat/internal/common"
	"github.com/M0hammadUsman/letschat/internal/domain"
	"os"
	"path/filepath"
	"runtime"
	"sync"
)

var (
	once             sync.Once
	client           *Client
	runProcessesOnce sync.Once
)

type Conversations map[string][]*domain.Message // key -> userID, val -> list of messages

// WsConnState will be switch cased by tui.TabContainerModel, so making it unique
type WsConnState int

const (
	Disconnected WsConnState = iota - 1
	WaitingForReconnection
	Reconnecting
	Connected
)

type WsConnStateChan chan WsConnState

type UsrLoggedIn bool
type UsrLoggedInChan chan UsrLoggedIn

type Client struct {
	// If zero valued -> requires login
	// then we set this AuthToken in the OS credential manager of respected Operating systems
	AuthToken string
	// it's where all the application related files will live on the client side from db, logging anything
	FilesDir string
	// will be used by tui.TabContainerModel
	ConnState WsConnState
	Repo      *repository.LocalRepository
	// currently logged-in user we fetch and populates this once there is a read from UsrLogin chan
	CurrentUsr *domain.User
	// once the user logs in we can read it from this channel anywhere in the application and do the required stuff
	// also tui will read the false case and redirects the user to login form
	// true -> successful login, false -> unauthorized requires login
	UsrLoggedIn UsrLoggedInChan
	// only write to this chan once there is successful login
	//SuccessfulUsrLogin chan struct{}
	// runs the tasks that needs a graceful shutdown, using BackgroundTask.Run
	BT *common.BackgroundTask
	// RunStartupProcesses runs long living processes, which dies on shutdown, some chores and
	// will be called from main method after there is write on RunningTui chan from tui.TabContainerModel
	// initialized in Init func
	RunStartupProcesses func()
	// talks to the api for managing native os based credential manager
	krm *keyringManager
	// for more info see Client.AttemptWsReconnectOnDisconnect, must be buffered
	wsConnStatus WsConnStateChan
	// every message will be written to this channel by websocket.Conn so we can read it from here
	messages domain.MsgChan
	// these are the conversations of the user displayed on the left side of the Conversations tab
	// once there is a ws conn they will be updated by current online status updates
	// if the connection is offline they will be fetched by the local db for offline view
	conversations Conversations
	// wrapper around *sqlx.DB
	db *repository.DB
	// directory to store application related files on client side, determined on startup for respected OS
	// supported OS -> windows, mac, linux
}

// Init initializes Storage Dirs, keyringManager to support access token storage at OS level,
// also opens a connection to sqlite DB, runs idempotent migrations, starts a goroutine to listen for user login,
// a goroutine to connect to Ws and listen for messages
// to get instance to a client use Get
func Init() error {
	var c Client
	var err error
	once.Do(func() {
		c.FilesDir, err = getAppStoragePath(appName)
		c.krm, err = newKeyringManager()
		if err != nil {
			return
		}
		c.AuthToken = c.krm.getAuthTokenFromKeyring()
		c.messages = make(domain.MsgChan, 16)
		c.BT = common.NewBackgroundTask()
		//c.UsrLoggedIn = true // we assume user is logged in, proved otherwise
		c.UsrLoggedIn = make(chan UsrLoggedIn)
		//c.SuccessfulUsrLogin = make(chan struct{})
		c.wsConnStatus = make(WsConnStateChan, 1)
		// Connecting to sqlite
		c.db, err = repository.OpenDB(c.FilesDir)
		if err != nil {
			return
		}
		c.Repo = repository.NewLocalRepository(c.db)
		// Running idempotent migrations
		err = c.db.RunMigrations()
	})
	if err != nil {
		if c.db != nil {
			_ = c.db.Close()
		}
		return err
	}
	c.RunStartupProcesses = sync.OnceFunc(func() {
		go c.ListenForUserLogin()
		go c.AttemptWsReconnectOnDisconnect()
		go c.WsConnectAndListenForMessages()
		u, err := c.Repo.GetCurrentUser()
		if err != nil {
			if errors.Is(err, domain.ErrRecordNotFound) {
				// as we cannot find the user in the db, user needs to log in again so
				// tui.TabContainerModel can redirect user to log-in page
				c.UsrLoggedIn <- false
			}
		}
		client.CurrentUsr = u
	})
	client = &c
	return nil
}

// Get returns singleton instance of a Client
func Get() *Client {
	return client
}

func (c *Client) ListenForMessages() domain.MsgChan {
	return c.messages
}

// Helpers & Stuff -----------------------------------------------------------------------------------------------------

func getAppStoragePath(appName string) (string, error) {
	var appStoragePath string
	homeDir, err := os.UserHomeDir() // Get the user's home directory
	if err != nil {
		return "", fmt.Errorf("could not find user home directory: %v", err)
	}

	switch opSys := runtime.GOOS; opSys {
	case "windows":
		// Windows: Store in C:/Users/username/AppData/Local/Programs/appName
		localAppData := os.Getenv("LOCALAPPDATA")
		if localAppData == "" {
			return "", fmt.Errorf("could not find LOCALAPPDATA environment variable")
		}
		appStoragePath = filepath.Join(localAppData, "Programs", appName)
	case "darwin":
		// macOS: Store in ~/Library/Application Support/appName
		appStoragePath = filepath.Join(homeDir, "Library", "Application Support", appName)
	case "linux":
		// Linux: Store in ~/.local/share/appName
		appStoragePath = filepath.Join(homeDir, ".local", "share", appName)
	default:
		return "", fmt.Errorf("unsupported OS: %s", opSys)
	}

	// Create the directory if it doesn't exist
	err = os.MkdirAll(appStoragePath, os.ModeDir)
	if err != nil {
		return "", fmt.Errorf("could not create app storage directory: %v", err)
	}

	return appStoragePath, nil
}
