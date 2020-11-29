package config

import (
	"fmt"

	"github.com/pkg/errors"

	"os"
	"reflect"
	"strings"

	"github.com/asaskevich/govalidator"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

// Config contains the required fields for running a server
type Config struct {
	LogLevel          string            `mapstructure:"LOG_LEVEL" valid:"required"`
	MRAssignerService MRAssignerService `mapstructure:"STATUS_SERVICE" valid:"required"`
}

type MRAssignerService struct {
	CommitHash  string `mapstructure:"COMMIT_HASH" valid:"required"`
	ServiceName string `mapstructure:"SERVICE_NAME" valid:"required"`
}

var CfgFile string

var CfgApp Config

func (c Config) Validate() error {
	_, err := govalidator.ValidateStruct(c)
	if err != nil {
		return errors.Wrap(err, "config: validate service config")
	}

	return nil
}

func InitConfig() {
	govalidator.SetFieldsRequiredByDefault(true)

	if CfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(CfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		// Search config in home directory with name ".filepath" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".config")
	}

	viper.AutomaticEnv() // read in environment variables that match

	replacer := strings.NewReplacer(".", "_")
	viper.SetEnvKeyReplacer(replacer)

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}

	bindEnvs(CfgApp)

	if err := viper.Unmarshal(&CfgApp); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if err := CfgApp.Validate(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func bindEnvs(iface interface{}, parts ...string) {
	ifv := reflect.ValueOf(iface)
	ift := reflect.TypeOf(iface)
	for i := 0; i < ift.NumField(); i++ {
		v := ifv.Field(i)
		t := ift.Field(i)
		tv, ok := t.Tag.Lookup("mapstructure")
		if !ok {
			continue
		}

		switch v.Kind() {
		case reflect.Struct:
			bindEnvs(v.Interface(), append(parts, tv)...)
		default:
			if err := viper.BindEnv(strings.Join(append(parts, tv), ".")); err != nil {
				os.Exit(1)
			}
		}
	}
}
