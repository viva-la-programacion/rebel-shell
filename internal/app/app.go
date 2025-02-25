package app

import "fmt"

// BlackhatApp is the core application struct.
type RebelshellApp struct {
	Config map[string]string
}

// NewBlackhatApp initializes the application.
func NewBlackhatApp() *RebelshellApp {
	return &RebelshellApp{
		Config: make(map[string]string),
	}
}

// Configure sets a configuration value.
func (app *RebelshellApp) Configure(key, value string) {
	app.Config[key] = value
	fmt.Printf("Configured %s = %s\n", key, value)
}

// GetConfig retrieves a configuration value.
func (app *RebelshellApp) GetConfig(key string) string {
	return app.Config[key]
}
