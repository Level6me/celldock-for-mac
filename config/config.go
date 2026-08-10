package config

import (
	"os"
	"path/filepath"
)

type Config struct {
	ListenAddr string
	DataDir    string
	DBPath     string
}

func LoadConfig() *Config {
	home, _ := os.UserHomeDir()
	dataDir := os.Getenv("DATA_DIR")
	if dataDir == "" {
		dataDir = filepath.Join(home, ".celldock")
	}
	os.MkdirAll(dataDir, 0755)

	listenAddr := os.Getenv("LISTEN_ADDR")
	if listenAddr == "" {
		listenAddr = ":8080"
	}

	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = filepath.Join(dataDir, "celldock.db")
	}

	return &Config{
		ListenAddr: listenAddr,
		DataDir:    dataDir,
		DBPath:     dbPath,
	}
}
