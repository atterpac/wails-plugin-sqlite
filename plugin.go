package sqlite

import (
	"database/sql"
	"errors"
	"runtime"

	"github.com/wailsapp/wails/v3/pkg/application"
)

// ---------------- Plugin Setup ----------------
// This is the main plugin struct. It can be named anything you like.
// It must implement the application.Plugin interface.
// Both the Init() and Shutdown() methods are called synchronously when the app starts and stops.

type Config struct {
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
	// Connection Options
	CacheShared        bool    // Default false
	MaxOpenConnections int     // Default 1
	MaxIdleConnections int     // Default 1
	Driver             string  // mattn/cgoless options
	DB                 *sql.DB // Connection created on Init
}

// Changing the name of this struct will change the name of the plugin in the frontend 
// Bound methods will exist inside frontend/bindings/sqlite/[PluginStruct]
type Sqlite struct{
    config *Config
    app *application.App
}

func NewPlugin(config *Config) *Sqlite {
	return &Sqlite{
        config: config,
    }
}

// Shutdown is called when the app is shutting down via runtime.Quit() call
// You can use this to clean up any resources you have allocated
func (p *Sqlite) Shutdown() {
	if p.config.DeleteOnShutdown {
		// Delete Database
	}
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
	if p.config.InMemory {
		return p.createMemDB()
	}
	if p.config.DbName == "" {
		return errors.New("Sqlite requires a DbName to be set or configured for In Memory database")
	}
	return p.createFileDB()
}

// ---------------- Plugin Methods ----------------
// Plugin methods are just normal Go methods. You can add as many as you like.
// The only requirement is that they are exported (start with a capital letter).
// You can also return any type that is JSON serializable.
// See https://golang.org/pkg/encoding/json/#Marshal for more information.
func (sqls Sqlite) Execute(cmd string, args... any) (int64, error) {
	result, err := sqls.config.DB.Exec(cmd, args)
	if err != nil {
		return 0, err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}
	return affected, nil
}

func (sqls Sqlite) Query(query string, args... any) (*sql.Rows, error) {
	return sqls.config.DB.Query(query, args)
}

func (sqls Sqlite) GetDB() (*sql.DB, error) {
	return sqls.config.DB, nil
}

func (sqls *Sqlite) SetDB(newDB *sql.DB) error {
	if newDB != nil {
		sqls.config.DB = newDB
		return nil
	}
	return errors.New("Attempted to set DB but provided pointer was nil")
}

func (sqls Sqlite) createMemDB() error {

	return nil
}

func (sqls Sqlite) createFileDB() error {
	switch runtime.GOOS {
	case "windows":
		// mkdir %APPDATA%
	case "darwin":
		// mkdir /ApplicationSupport/AppData
	case "linux":
		// mkdir XDG DataDir
	default:
		return errors.New("Operating system not supported, please use Windows/MacOs/Linux")
	}
	// Create Direcotry if doesnt exist
	// Create sqlite DB with name inside of Dir
	return nil
}
