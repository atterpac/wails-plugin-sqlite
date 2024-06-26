package sqlite

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	_ "github.com/mattn/go-sqlite3"
	"github.com/wailsapp/wails/v3/pkg/application"
)

// ---------------- Plugin Setup ----------------
// This is the main plugin struct. It can be named anything you like.
// It must implement the application.Plugin interface.
// Both the Init() and Shutdown() methods are called synchronously when the app starts and stops.
// Changing the name of this struct will change the name of the plugin in the frontend
// Bound methods will exist inside frontend/bindings/sqlite/[PluginStruct]
type Sqlite struct {
	// Add any configuration options here
	DbName string // Required input
	// Memory Database
	InMemory bool // Default false
	// File Locations
	MacDir     string // XDG Default/DbName/DbName.db
	WindowsDir string // XDG Default/DbName/DbName.db
	LinuxDir   string // XDG Default/DbName/DbName.db
	// Shutdown Options
	DeleteOnShutdown bool // Default false
	DeleteDir        bool // Default false
	// Connection Options
	CacheShared        bool    // Default false
	MaxOpenConnections int     // Default 1
	MaxIdleConnections int     // Default 2
	DB                 *sql.DB // Connection created on Init
	savedPath          string  // interal use for cleanup
	app    *application.App
}

// Shutdown is called when the app is shutting down via runtime.Quit() call
// You can use this to clean up any resources you have allocated
func (p *Sqlite) Shutdown() error {
	if p.DeleteOnShutdown {
		// Delete Database
		if p.InMemory {
			return nil
		}
		var err error
		if p.DeleteDir {
			err = os.RemoveAll(filepath.Dir(p.savedPath))
		} else {
			err = os.RemoveAll(p.savedPath)
		}
		if err != nil {
			return fmt.Errorf("failed to delete database: %w", err)
		}
	}
	return nil
}

// Name returns the name of the plugin.
// You should use the go module format e.g. github.com/myuser/myplugin
func (p *Sqlite) Name() string {
	return "github.com/atterpac/wails-plugin-sqlite"
}

// Init is called when the app is starting up. You can use this to
// initialise any resources you need. You can also access the application
// instance via the app property.
func (p *Sqlite) Init() error {
	p.app = application.Get()
	if p.InMemory {
		return p.createMemDB()
	}
	if p.DbName == "" {
		return errors.New("Sqlite requires a DbName to be set or configured for In Memory database")
	}
	return p.createFileDB()
}

// ---------------- Plugin Methods ----------------
// Plugin methods are just normal Go methods. You can add as many as you like.
// The only requirement is that they are exported (start with a capital letter).
// You can also return any type that is JSON serializable.
// See https://golang.org/pkg/encoding/json/#Marshal for more information.
func (sqls Sqlite) Execute(cmd string, args ...any) (int64, error) {
	result, err := sqls.DB.Exec(cmd, args)
	if err != nil {
		return 0, err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}
	return affected, nil
}

func (sqls Sqlite) Query(query string, args ...any) (*sql.Rows, error) {
	return sqls.DB.Query(query, args)
}

func (sqls Sqlite) GetDB() (*sql.DB, error) {
	return sqls.DB, nil
}

func (sqls *Sqlite) SetDB(newDB *sql.DB) error {
	if newDB != nil {
		sqls.DB = newDB
		return nil
	}
	return errors.New("Attempted to set DB but provided pointer was nil")
}

func (sqls *Sqlite) createMemDB() error {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		return fmt.Errorf("failed to create in-memory database: %w", err)
	}

	sqls.DB = db
	return nil
}

func (sqls *Sqlite) createFileDB() error {
	var dbPath string
	switch runtime.GOOS {
	case "windows":
		if sqls.WindowsDir != "" {
			dbPath = sqls.WindowsDir
		} else {
			dbPath = filepath.Join(os.Getenv("APPDATA"), sqls.WindowsDir)
		}
	case "darwin":
		if sqls.MacDir != "" {
			dbPath = sqls.MacDir
		} else {
			dbPath = filepath.Join(os.Getenv("HOME"), "Library", "Application Support", sqls.DbName, sqls.DbName+".db")
		}

	case "linux":
		if sqls.LinuxDir != "" {
			dbPath = sqls.LinuxDir
		} else {
			dbPath = filepath.Join(os.Getenv("HOME"), ".config", sqls.DbName)
		}
	default:
		return errors.New("operating system not supported, please use Windows/macOS/Linux")
	}
	fileName := sqls.DbName + ".db"

	if sqls.CacheShared {
		fileName = fileName + "?cache=shared"
	}

	// Create directory if it doesn't exist
	err := os.MkdirAll(dbPath, 0755)
	if err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Create sqlite DB with name inside of Dir
	dbPath = filepath.Join(dbPath, fileName)
	sqls.savedPath = fileName
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return fmt.Errorf("failed to create database: %w", err)
	}

	idle := sqls.MaxIdleConnections
	open := sqls.MaxOpenConnections
	// Set connection options
	if idle == 0 {
		idle = 2
	}
	if open == 0 {
		open = 1
	}
	db.SetMaxIdleConns(idle)
	db.SetMaxOpenConns(open)

	err = db.Ping()
	if err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	sqls.app.Logger.Info("Sqlite Initialized", "path", dbPath, "idle", idle, "open", open)
	sqls.DB = db
	return nil
}
