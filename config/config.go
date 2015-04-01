package config

import (
	"os"
	"flag"
)

type Config struct {
	ListenPort string `toml:"port"`
	EtcdAddr string `toml:"etcd-addr"`
}

func New() *Config {
	c := new(Config)
	c.ListenPort = "8080"
	c.EtcdAddr = "localhost:4001"
	return c
}

// Loads configuration from command line flags.
func (c *Config) LoadFlags(arguments []string) error {
	
	f := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	f.StringVar(&c.ListenPort, "port", c.ListenPort, "")
	f.StringVar(&c.EtcdAddr, "etcd-addr", c.EtcdAddr, "")

	if err := f.Parse(arguments); err != nil {
		return err
	}	
	
	return nil
}
