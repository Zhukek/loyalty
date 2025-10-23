package main

import (
	"flag"

	"github.com/caarlos0/env"
)

type Config struct {
	Address        string `env:"RUN_ADDRESS"`
	DBURI          string `env:"DATABASE_URI"`
	AccrualAddress string `env:"ACCRUAL_SYSTEM_ADDRESS"`
}

func getParams() *Config {
	const (
		defaultAddress = "localhost:8080"
		defaultDBURI   = ""
		defaultAccrual = ""
	)

	config := Config{}

	flag.StringVar(&config.Address, "a", defaultAddress, "address:port")
	flag.StringVar(&config.DBURI, "d", defaultDBURI, "database URI")
	flag.StringVar(&config.AccrualAddress, "r", defaultAccrual, "accrual system address")

	flag.Parse()

	env.Parse(&config)

	return &config
}
