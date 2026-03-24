package config

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type Config struct {
	DbURL           string `json:"db_url"`
	CurrentUserName string `json:"current_user_name"`
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
		fmt.Println(err)
	}
	defer file.Close()
	data, err := io.ReadAll(file)
	if err != nil {
		return &configStruct, err
	}
	//fmt.Printf("raw data %v\n", data)
	if err := json.Unmarshal(data, &configStruct); err != nil {
		return &configStruct, err
	}
	//fmt.Printf("unmarshalled data %v", configStruct)
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
func (c *Config) SetUser(userName string) error {
	c.CurrentUserName = userName
	//fmt.Printf("updated struct, %v\n", c)
	if err := write(c); err != nil {
		return err
	}

	return nil
}

func write(cfg *Config) error {
	fmt.Println("writing file")
	//fmt.Println(cfg)
	path, err := getConfigFilePath()
	//fmt.Println(path)
	if err != nil {
		return err
	}
	jsonData, err := json.Marshal(cfg)
	//fmt.Printf("Marshelled data, %v", jsonData)
	/*file, err := os.Open(path)
	if err != nil {
		fmt.Println("difficulty opening file")
		return err
	}
	defer file.Close()*/

	err = os.WriteFile(path, jsonData, 0666)
	if err != nil {
		fmt.Println("difficulty writing file")
		return err
	}
	return nil
}
