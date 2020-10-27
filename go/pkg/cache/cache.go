package cache

import (
	"encoding/json"
	"fmt"

	"github.com/GitbookIO/diskache"
	"github.com/mitchellh/go-homedir"

	"github.com/replicate/replicate/go/pkg/console"
)

var instance *Cache = nil

type Cache struct {
	dc *diskache.Diskache
}

func Instance() (*Cache, error) {
	if instance != nil {
		return instance, nil
	}
	dir, err := homedir.Expand("~/.cache/replicate")
	if err != nil {
		return nil, fmt.Errorf("Failed to determine home directory: %s", err)
	}
	dc, err := diskache.New(&diskache.Opts{
		Directory: dir,
	})
	if err != nil {
		return nil, fmt.Errorf("Failed to create disk cache: %s", err)
	}
	instance = &Cache{dc: dc}
	return instance, nil
}

func Set(key string, data []byte) error {
	c, err := Instance()
	if err != nil {
		return err
	}
	if err := c.dc.Set(key, data); err != nil {
		return fmt.Errorf("Failed to store %s in cache: %s", key, err)
	}
	return nil
}

func Get(key string) (data []byte, ok bool) {
	c, err := Instance()
	if err != nil {
		console.Warn("Failed to get cache instance: %s", err)
		return nil, false
	}
	return c.dc.Get(key)
}

func SetString(key string, s string) error {
	return Set(key, []byte(s))
}

func GetString(key string) (s string, ok bool) {
	data, ok := Get(key)
	if !ok {
		return "", false
	}
	return string(data), true
}

func SetStruct(key string, v interface{}) error {
	data, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("Failed to marshal cache data: %s", err)
	}
	return Set(key, data)
}

func GetStruct(key string, v interface{}) bool {
	data, ok := Get(key)
	if !ok {
		return false
	}
	err := json.Unmarshal(data, v)
	if err != nil {
		console.Warn("Failed to unmarshal cached data: %s", err)
		return false
	}
	return true
}
