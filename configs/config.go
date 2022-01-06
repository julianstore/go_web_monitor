package configs

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/webmonitor/web-monitor/constants"
)

// Config contains configuration of all components of the WalletNode.
type Config struct {
	Main `json:","`
	Init `json:","`
}

// String : returns string from Config fields
func (config *Config) String() (string, error) {
	// The main purpose of using a custom converting is to avoid unveiling credentials.
	// All credentials fields must be tagged `json:"-"`.
	data, err := json.Marshal(config)

	if err != nil {
		return "", err
	}

	return string(data), nil
}

// LoadConfig : load the config from config file
func LoadConfig() (cofig *Config, err error) {
	data, err := ioutil.ReadFile(constants.ConfigFile)

	if err != nil {
		return nil, err
	}

	var dataConf Config
	err = json.Unmarshal(data, &dataConf)

	return &dataConf, err
}

// New returns a new Config instance
func New() *Config {
	return &Config{
		Main: *NewMain(),
	}
}

// GetConfig : Get the config from config file. If there is no config file then create a new config.
func GetConfig() *Config {
	var config *Config
	var err error
	if CheckFileExist(constants.ConfigFile) {
		config, err = LoadConfig()
		if err != nil {
			fmt.Println("FILE IS MISSING")
			os.Exit(-1)
		}
	} else {
		config = New()
	}
	return config
}

// CheckFileExist check the file exist
func CheckFileExist(filepath string) bool {
	var err error
	if _, err = os.Stat(filepath); os.IsNotExist(err) {
		return false
	}
	return true
}


// SaveConfig : save config
func (config *Config) SaveConfig() error {
	data, err := config.String()

	if err != nil {
		return err
	}

	if ioutil.WriteFile(constants.ConfigFile, []byte(data), 0644) != nil {
		return err
	}
	return nil
}