package utils

import (
	"fmt"
	"os"
)

const (
	EnvAddress = "FTP_ADDR"
	EnvUser    = "FTP_USER"
	EnvPass    = "FTP_PASS"
)

type Config struct {
	Address  string
	User     string
	Password string
}

func (c *Config) Validate() error {
	if c.Address == "" {
		return fmt.Errorf("FTP server address is not set")
	}
	if c.User == "" {
		return fmt.Errorf("FTP server user account is not set")
	}
	if c.Password == "" {
		return fmt.Errorf("FTP server account password is not set")
	}
	return nil
}

func LoadConfig() (*Config, error) {
	c := &Config{
		Address:  os.Getenv(EnvAddress),
		User:     os.Getenv(EnvUser),
		Password: os.Getenv(EnvPass),
	}

	if err := c.Validate(); err != nil {
		return nil, err
	}
	return c, nil
}
