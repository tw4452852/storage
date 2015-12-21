package storage

import (
	"encoding/xml"
	"log"
	"os"
	"path/filepath"
)

type Config struct { /*{{{*/
	Type     string `xml:"type"`
	Root     string `xml:"root"`
	User     string `xml:"username"`
	Password string `xml:"password"`
} /*}}}*/

type Configs struct { /*{{{*/
	Content []*Config `xml:"repo"`
} /*}}}*/

func getConfig(path string) (*Configs, error) { /*{{{*/
	file, err := os.Open(path)
	if err != nil {
		log.Printf("open config file error: %s\n", err)
		return nil, err
	}
	defer file.Close()
	decoder := xml.NewDecoder(file)
	cfg := new(Configs)
	if err := decoder.Decode(cfg); err != nil {
		log.Printf("parse config file error: %s\n", err)
		return nil, err
	}

	cs := make([]*Config, 0, len(cfg.Content))
	for _, c := range cfg.Content {
		// filter the empty repo
		if c.Type == "" || c.Root == "" {
			continue
		}
		// join the $GOPATH to the rel local root path
		if c.Type == "local" {
			c.Root = filepath.FromSlash(c.Root)
			//TODO: windows isabs not begin with '/'
			if !filepath.IsAbs(c.Root) {
				c.Root = filepath.Join(os.Getenv("GOPATH"), c.Root)
			}
		}
		cs = append(cs, c)
	}

	cfg.Content = cs
	return cfg, nil
} /*}}}*/
