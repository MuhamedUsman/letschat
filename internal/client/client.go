package client

import (
	"context"
	"errors"
	"fmt"
	"github.com/MuhamedUsman/letschat/internal/client/repository"
	"github.com/MuhamedUsman/letschat/internal/common"
	"github.com/MuhamedUsman/letschat/internal/domain"
	"github.com/coder/websocket"
	"os"
	"path/filepath"
	"runtime"
	"sync"
)

var (
	once   sync.Once
	client *Client
)

type Client struct {
	// If zero valued -> requires login
	// then we set this AuthToken in the OS credential manager of respected Operating systems
	AuthToken string
	// it's where all the application related files will live on the client side from db, logging anything
	FilesDir string
	// currently logged-in user we fetch and populates this once there is a read from UsrLogin chan
	CurrentUsr  *domain.User
	WsConnState *WsConnBroadcaster
	LoginState  *LoginBroadcaster
	RecvMsgs    *RecvMsgsBroadcaster
	// these are the conversations of the user displayed on the left side of the Conversations tab
	// once there is a ws conn they will be updated by current online status updates
	// if the connection is offline they will be fetched by the local db for offline view
	// we've to wait for state change to get updated conversations, they will be auto updated on WsConnState changes
	Conversations *ConvosBroadcaster
	// runs the tasks that needs a graceful shutdown, using BackgroundTask.Run
	BT *common.BackgroundTask
	// RunStartupProcesses runs long living processes, which dies on shutdown, some chores and
	// will be called from main method after there is write on RunningTui chan from tui.TabContainerModel
	// initialized in Init func
	RunStartupProcesses func()
	wsConn              *websocket.Conn
	// talks to the api for managing native os based credential manager
	krm      *keyringManager
	sentMsgs sentMsgs
	// wrapper around *sqlx.DB
	db *repository.DB
	// directory to store application related files on client side, determined on startup for respected OS
	// supported OS -> windows, mac, linux
	repo *repository.LocalRepository
}

// Init initializes Storage Dirs, keyringManager to support access token storage at OS level,
// also opens a connection to sqlite DB, runs idempotent migrations, starts a goroutine to listen for user login,
// a goroutine to connect to Ws and listen for recvMsgs
// to get instance to a client use Get
func Init(key int) error {
	var c Client
	var err error
	once.Do(func() {
		c.FilesDir, err = getAppStoragePath(appName)
		c.krm, err = newKeyringManager(key)
		if err != nil {
			return
		}
		c.AuthToken = c.krm.getAuthTokenFromKeyring()
		c.BT = common.NewBackgroundTask()
		c.WsConnState = newWsConnBroadcaster()
		c.LoginState = newLoginBroadcaster()
		c.Conversations = newConvosBroadcaster()
		c.RecvMsgs = newRecvMsgsBroadcaster()
		// Connecting to sqlite
		c.db, err = repository.OpenDB(c.FilesDir, key)
		if err != nil {
			return
		}
		c.repo = repository.NewLocalRepository(c.db)
		// Running idempotent migrations
		err = c.db.RunMigrations()
	})
	if err != nil {
		return err
	}
	c.RunStartupProcesses = sync.OnceFunc(func() {
		c.BT.Run(func(shtdwnCtx context.Context) { c.LoginState.Broadcast(shtdwnCtx) })
		c.BT.Run(func(shtdwnCtx context.Context) { c.WsConnState.Broadcast(shtdwnCtx) })
		c.BT.Run(func(shtdwnCtx context.Context) { c.Conversations.Broadcast(shtdwnCtx) })
		c.BT.Run(func(shtdwnCtx context.Context) { c.RecvMsgs.Broadcast(shtdwnCtx) })
		c.BT.Run(func(shtdwnCtx context.Context) { c.handleReceivedMsgs(shtdwnCtx) })
		c.BT.Run(func(shtdwnCtx context.Context) { c.manageUserLogins(shtdwnCtx) })
		c.BT.Run(func(shtdwnCtx context.Context) { c.attemptWsReconnectOnDisconnect(shtdwnCtx) })
		c.BT.Run(func(shtdwnCtx context.Context) { c.wsConnectAndListenForMessages(shtdwnCtx) })
		c.BT.Run(func(shtdwnCtx context.Context) { c.populateConversationsAccordingToWsConnState(shtdwnCtx) })
		u, err := c.repo.GetCurrentUser()
		if err != nil {
			if errors.Is(err, domain.ErrRecordNotFound) {
				// as we cannot find the user in the db, user needs to log in again so
				// tui.TabContainerModel can redirect user to log-in page
				c.LoginState.Write(false)
			}
		}
		c.CurrentUsr = u
	})
	client = &c
	return nil
}

// Get returns singleton instance of a Client
func Get() *Client {
	return client
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
