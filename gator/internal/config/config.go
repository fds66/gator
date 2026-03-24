package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Config struct {
	DbURL           string `json:"db_url"`
	CurrentUserName string `json:"current_user_name"`
}

func (c *Config) SetUser(userName string) error {
	c.CurrentUserName = userName
	//fmt.Printf("updated struct, %v\n", c)
	if err := write(c); err != nil {
		return err
	}

	return nil
}

/* JSON structure
{
  "db_url": "connection_string_goes_here",
  "current_user_name": "username_goes_here"
}
*/

const configFileName = ".gatorconfig.json"

func Read() (*Config, error) {
	configStruct := Config{
		DbURL:           "",
		CurrentUserName: "",
	}
	json_config_path, err := getConfigFilePath()
	if err != nil {
		return &configStruct, err
	}
	file, err := os.Open(json_config_path)
	if err != nil {
		return &configStruct, err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	err = decoder.Decode(&configStruct)
	if err != nil {
		return &configStruct, err
	}

	return &configStruct, nil
}
func getConfigFilePath() (string, error) {

	homeDirectory, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	json_config_path := filepath.Join(homeDirectory, configFileName)
	//fmt.Println(json_config_path)
	return json_config_path, nil
}

func write(cfg *Config) error {
	fmt.Println("writing file")
	//fmt.Println(cfg)
	path, err := getConfigFilePath()
	//fmt.Println(path)
	if err != nil {
		return err
	}
	file, err := os.Create(path)

	if err != nil {
		return err
	}
	defer file.Close()
	encoder := json.NewEncoder(file)
	err = encoder.Encode(cfg)
	if err != nil {
		return err
	}
	return nil
}
