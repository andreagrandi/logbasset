package app

const (
	Name    = "logbasset"
	Version = "0.2.1"
	Author  = "Andrea Grandi"
	License = "Apache-2.0"
)

type App struct {
	Name    string
	Version string
	Author  string
	License string
}

func New() *App {
	return &App{
		Name:    Name,
		Version: Version,
		Author:  Author,
		License: License,
	}
}

func (a *App) GetFullVersion() string {
	return a.Name + " version " + a.Version
}
