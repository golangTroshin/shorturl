package config

import "flag"

var Options struct {
	FlagRunAddr          string
	FlagBaseShortURLAddr string
}

func ParseFlags() {
	flag.StringVar(&Options.FlagRunAddr, "a", "localhost:8080", "address and port to run server")
	flag.StringVar(&Options.FlagBaseShortURLAddr, "b", "http://localhost:8080", "base result url")

	flag.Parse()
}
