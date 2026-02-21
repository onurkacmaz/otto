package db

import (
	"encoding/json"
	"os"
	"path/filepath"
)

const historyDir = ".otto"
const historyFile = "history.json"

func historyPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, historyDir, historyFile)
}

func DisplayName(cfg Config) string {
	if cfg.Name != "" {
		return cfg.Name
	}
	host := cfg.Host
	if host == "" {
		host = "localhost"
	}
	user := cfg.User
	dbname := cfg.DBName

	port := cfg.Port
	if cfg.Driver == DriverMySQL {
		if port == "" {
			port = "3306"
		}
		if user == "" {
			user = "root"
		}
	} else {
		if port == "" {
			port = "5432"
		}
		if user == "" {
			user = "postgres"
		}
		if dbname == "" {
			dbname = "postgres"
		}
	}
	return user + "@" + host + ":" + port + "/" + dbname
}

func matchKey(a, b Config) bool {
	return DisplayName(a) == DisplayName(b)
}

func LoadHistory() []Config {
	data, err := os.ReadFile(historyPath())
	if err != nil {
		return nil
	}
	var history []Config
	if err := json.Unmarshal(data, &history); err != nil {
		return nil
	}
	return history
}

func SaveConnection(cfg Config) {
	history := LoadHistory()

	found := false
	for i, h := range history {
		if matchKey(h, cfg) {
			history[i] = cfg
			found = true
			break
		}
	}
	if !found {
		history = append([]Config{cfg}, history...)
	}

	p := historyPath()
	_ = os.MkdirAll(filepath.Dir(p), 0700)

	data, err := json.MarshalIndent(history, "", "  ")
	if err != nil {
		return
	}
	_ = os.WriteFile(p, data, 0600)
}
