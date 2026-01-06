package main

import (
	"context"
	"encoding/json"
	"github.com/wailsapp/wails/v2/pkg/runtime"
	"log"
	"net/http"
	"strings"
)

type Storer interface {
	Open(dbPath, decryptKey, compression string) (err error)
	Set(key string, value []byte) error
	Get(key string) ([]byte, error)
	Delete(key string) error
	List(limit *int, startCursor *string) (keys []string, cursor string, err error)
	Search(prefix string, limit *int, offset int) (keys []string, err error)
	IsRunning() bool
	Close()
}

type messageType string

const (
	TypeOpen   messageType = "open"
	TypeSet    messageType = "set"
	TypeDelete messageType = "delete"
	TypeList   messageType = "list"
	TypeGet    messageType = "get"
	TypeSearch messageType = "search"

	OkResponse                 = "ok"
	NotRunningResponse         = "db isn't running"
	UnknownMessageTypeResponse = "unknown message type"
)

type AppMessage struct {
	Type messageType `json:"type"`
	Body string      `json:"body"`
}

type MessageOpen struct {
	Path          string `json:"path"`
	DecryptionKey string `json:"decryption_key"`
	Compression   string `json:"compression"`
	Delimiter     string `json:"delimiter"`
}

type MessageSet struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type MessageDelete struct {
	Key string `json:"key"`
}

type MessageGet MessageDelete

type MessageList struct {
	Limit  *int    `json:"limit"`
	Cursor *string `json:"cursor"`
}

type MessageSearch struct {
	Prefix string `json:"prefix"`
	Limit  *int   `json:"limit"`
	Offset int    `json:"offset"`
}

type ListResponse struct {
	Cursor string   `json:"cursor"`
	Keys   []string `json:"keys"`
}

type SearchResponse struct {
	Keys   []string `json:"keys"`
	Offset int      `json:"offset"`
}

type Item struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type App struct {
	ctx       context.Context
	db        Storer
	delimiter string
}

// NewApp creates a new App application struct
func NewApp(db Storer) *App {
	return &App{db: db}
}

// Startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	log.Println("starting application")
}

// OpenDirectoryDialog opens a directory picker dialog
func (a *App) OpenDirectoryDialog() string {
	path, err := runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Select Badger Database Folder",
	})
	if err != nil {
		log.Printf("error opening directory dialog: %v", err)
		return ""
	}
	return path
}

// Call calls a JS/Go mapped method
func (a *App) Call(msg AppMessage) (response AppMessage) {
	// Log message type without exposing sensitive data
	log.Printf("received message type: %s", msg.Type)

	switch msg.Type {
	case TypeOpen:
		if a.db.IsRunning() {
			log.Printf("database already running")
			return AppMessage{msg.Type, "already running"}
		}
		var openMsg MessageOpen
		if err := json.Unmarshal([]byte(msg.Body), &openMsg); err != nil {
			log.Printf("unmarshaling open message: %v", err)
			return AppMessage{msg.Type, err.Error()}
		}

		log.Printf("Opening database at path: %s, compression: %s", openMsg.Path, openMsg.Compression)
		if err := a.db.Open(openMsg.Path, openMsg.DecryptionKey, openMsg.Compression); err != nil {
			log.Printf("opening database: %v", err)
			return AppMessage{msg.Type, err.Error()}
		}
		a.delimiter = openMsg.Delimiter
		log.Printf("Database opened successfully with delimiter: %s", a.delimiter)
		return AppMessage{msg.Type, OkResponse}
	case TypeSet:
		if !a.db.IsRunning() {
			log.Printf("Database not running for set operation")
			return AppMessage{msg.Type, NotRunningResponse}
		}
		var setMsg MessageSet
		if err := json.Unmarshal([]byte(msg.Body), &setMsg); err != nil {
			log.Printf("unmarshaling set message: %v", err)
			return AppMessage{msg.Type, err.Error()}
		}
		if err := a.db.Set(setMsg.Key, []byte(setMsg.Value)); err != nil {
			log.Printf("setting key %s: %v", setMsg.Key, err)
			return AppMessage{msg.Type, err.Error()}
		}
		log.Printf("key %s set successfully", setMsg.Key)
		return AppMessage{msg.Type, OkResponse}
	case TypeGet:
		if !a.db.IsRunning() {
			log.Printf("database not running for get operation")
			return AppMessage{msg.Type, NotRunningResponse}
		}
		var getMsg MessageGet
		if err := json.Unmarshal([]byte(msg.Body), &getMsg); err != nil {
			log.Printf("unmarshaling get message: %v", err)
			return AppMessage{msg.Type, err.Error()}
		}
		value, err := a.db.Get(getMsg.Key)
		if err != nil {
			log.Printf("getting key %s: %v", getMsg.Key, err)
			return AppMessage{msg.Type, err.Error()}
		}
		log.Printf("key %s retrieved successfully, value length: %d", getMsg.Key, len(value))
		if isImage(value) {
			value = []byte("[image]")
		}
		bt, _ := json.Marshal(Item{Key: getMsg.Key, Value: string(value)})
		return AppMessage{msg.Type, string(bt)}
	case TypeDelete:
		if !a.db.IsRunning() {
			log.Printf("Database not running for delete operation")
			return AppMessage{msg.Type, NotRunningResponse}
		}
		var deleteMsg MessageDelete
		if err := json.Unmarshal([]byte(msg.Body), &deleteMsg); err != nil {
			log.Printf("unmarshaling delete message: %v", err)
			return AppMessage{msg.Type, err.Error()}
		}
		if err := a.db.Delete(deleteMsg.Key); err != nil {
			log.Printf("deleting key %s: %v", deleteMsg.Key, err)
			return AppMessage{msg.Type, err.Error()}
		}
		log.Printf("key %s deleted successfully", deleteMsg.Key)
		return AppMessage{msg.Type, OkResponse}
	case TypeList:
		if !a.db.IsRunning() {
			log.Printf("database not running for list operation")
			return AppMessage{msg.Type, NotRunningResponse}
		}
		var listMsg MessageList
		if err := json.Unmarshal([]byte(msg.Body), &listMsg); err != nil {
			log.Printf("unmarshaling list message: %v", err)
			return AppMessage{msg.Type, err.Error()}
		}
		keys, cursor, err := a.db.List(listMsg.Limit, listMsg.Cursor)
		if err != nil {
			log.Printf("listing items failure: %v", err)
		}
		bt, _ := json.Marshal(ListResponse{Cursor: cursor, Keys: keys})
		log.Printf("Listed %d items, cursor: %s", len(keys), cursor)
		return AppMessage{msg.Type, string(bt)}
	case TypeSearch:
		if !a.db.IsRunning() {
			log.Printf("database not running for list operation")
			return AppMessage{msg.Type, NotRunningResponse}
		}
		var searchMsg MessageSearch
		if err := json.Unmarshal([]byte(msg.Body), &searchMsg); err != nil {
			log.Printf("unmarshaling list message: %v", err)
			return AppMessage{msg.Type, err.Error()}
		}

		keys, err := a.db.Search(searchMsg.Prefix, searchMsg.Limit, searchMsg.Offset)
		if err != nil {
			log.Printf("listing items failure: %v", err)
		}
		bt, _ := json.Marshal(SearchResponse{Keys: keys, Offset: len(keys)})
		log.Printf("Found %d items", len(keys))
		return AppMessage{msg.Type, string(bt)}
	default:
		log.Printf("unsupported message type: %s", msg.Type)
		return AppMessage{"", UnknownMessageTypeResponse}
	}
}

func (a *App) close(_ context.Context) {
	a.db.Close()
	log.Println("app closed")
}

func isImage(data []byte) bool {
	contentType := http.DetectContentType(data)
	switch contentType {
	case "image/png":
		return true
	case "image/jpeg":
		return true
	case "image/gif":
		return true
	case "image/webp":
		return true
	default:
		if strings.Contains(string(data), "data:image") {
			return true
		}
		return false
	}
}
