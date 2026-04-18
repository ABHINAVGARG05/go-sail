package initializers

import (
	"log"
	"os"

	"gopkg.in/yaml.v3"

	"github.com/TejasGhatte/go-sail/internal/models"
)

var Config models.Config

func LoadConfig(filename string) {
	data, err := os.ReadFile(filename)
	if err != nil {
		log.Fatalf("error reading config file: %v", err)
	}

	err = yaml.Unmarshal(data, &Config)
	if err != nil {
		log.Fatalf("error parsing config file: %v", err)
	}
}
