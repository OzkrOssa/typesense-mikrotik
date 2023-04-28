package utils

import (
	"log"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Host []string `yaml:"host"`
}

func LoadConfig() []string {

	fileName, _ := filepath.Abs("mikrotik.yml")
	data, err := os.ReadFile(fileName)
	if err != nil {
		log.Println(err)
	}

	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		panic(err)
	}

	return config.Host
}
