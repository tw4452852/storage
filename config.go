package storage

import (
	"encoding/json"
	"log"
	"os"
)

type Config struct {
	Type     string `json:"type"`
	Root     string `json:"root"`
	User     string `json:"username"`
	Password string `json:"password"`
}

type Configs []*Config

func getConfig(path string) (Configs, error) {
	file, err := os.Open(path)
	if err != nil {
		log.Printf("open config file error: %s\n", err)
		return nil, err
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	cfg := make(Configs, 0)
	if err := decoder.Decode(&cfg); err != nil {
		log.Printf("parse config file error: %s\n", err)
		return nil, err
	}

	cs := make([]*Config, 0, len(cfg))
	for _, c := range cfg {
		// filter the empty repo
		if c.Type == "" || c.Root == "" {
			continue
		}
		cs = append(cs, c)
	}

	return cs, nil
}
