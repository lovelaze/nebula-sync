package config

import (
	"github.com/joho/godotenv"
	"github.com/rs/zerolog/log"
)

func LoadEnvFile(filename string) error {
	log.Debug().Msgf("Loading envs from file: %s", filename)
	return godotenv.Load(filename)
}
