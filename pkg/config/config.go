package config

import (
	"log"
	"os"
	"path/filepath"
	"sync"

	"gopkg.in/yaml.v3"
)

var instance *Config

var once sync.Once

type Config struct {
	Port           uint   `yaml:"Port"`
	DBPath         string `yaml:"OMSDBPath"`
	RepositoryPath string `yaml:"RepositoryPath"` // Export path for the files

}

func GetConfig() *Config {
	once.Do(func() {
		err := initConfig()
		if err != nil {
			log.Fatalf("[config] initialization failed - error: %s", err.Error())
		}
	})

	return instance
}

func initConfig() error {
	if _, err := os.Stat("./config/config.yaml"); err != nil {
		err = createConfig()
		if err != nil {
			return err
		}
	}

	file, err := os.Open("./config/config.yaml")
	if err != nil {
		return err
	}
	defer file.Close()

	d := yaml.NewDecoder(file)

	if err := d.Decode(&instance); err != nil {
		return err
	}

	return nil
}

func createConfig() error {
	config := Config{
		Port:           8888,
		DBPath:         "",
		RepositoryPath: "./repository",
	}

	data, err := yaml.Marshal(&config)
	if err != nil {
		return err
	}

	_, err = os.Stat("./config")
	if os.IsNotExist(err) {
		err = os.Mkdir("./config", os.ModePerm)
		if err != nil {
			return err
		}
	}

	err = os.WriteFile("./config/config.yaml", data, 0600)
	if err != nil {
		return err
	}

	pth, err := filepath.Abs("./config/config.yaml")
	if err != nil {
		log.Printf("[config] could not get absolute path")
		pth = "./config/config.yaml"
	}

	log.Printf("[config] created config.yaml path: %s", pth)
	return nil
}
