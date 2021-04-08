package config

import "github.com/spf13/viper"

type Config struct {
	ArangoHost     string `json:"arango_host"`
	ArangoUser     string `json:"arango_user"`
	ArangoPassword string `json:"arango_password"`
	Secret         string `json:"secret"`
}

var Conf Config

func init() {
	viper.SetConfigName("config") // name of config file (without extension)
	viper.SetConfigType("json")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}

	err = viper.Unmarshal(&Conf)
	if err != nil {
		panic(err)
	}
}
