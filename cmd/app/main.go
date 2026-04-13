package main

import (
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
	"github.com/pardnchiu/ThreadsMarketing/internal/tui"
	utils_keychain "github.com/pardnchiu/go-utils/filesystem/keychain"
)

const (
	serviceName = "ThreadsMarketing"
)

var (
	configDir = filepath.Join(os.Getenv("HOME"), ".config", serviceName)
)

func main() {
	_ = godotenv.Load()

	utils_keychain.Init(serviceName, configDir)

	app := tui.New()
	if err := app.Run(); err != nil {
		panic(err)
	}
}
