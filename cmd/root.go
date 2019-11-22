/*
Copyright © 2019 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/king526/goupdater/sdk"
	"github.com/spf13/cobra"

	"github.com/spf13/viper"
)

var (
	cfgFile    string
	filter     string
	server     string
	serverList []sdk.ServerEntity
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "goupdater",
	Short: "A brief description of your application",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file")
	rootCmd.PersistentFlags().StringVar(&filter, "filter", "", "server filters")
	rootCmd.PersistentFlags().StringVarP(&server, "server", "s", "", "server special(host:port) split by ,")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.

}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile == "" {
		for _, s := range strings.Split(server, ",") {
			if s = strings.TrimSpace(s); s != "" {
				serverList = append(serverList, sdk.ServerEntity{Addr: s})
			}
		}

	} else {
		viper.SetConfigFile(cfgFile)
		// If a config file is found, read it in.
		if err := viper.ReadInConfig(); err != nil {
			fmt.Println("ReadInConfig:", err)
			os.Exit(0)
		}
		if err := viper.UnmarshalKey("server", &serverList); err != nil {
			fmt.Println("parse server failed:", err)
			os.Exit(0)
		}
	}
	if len(serverList) == 0 {
		fmt.Println("no target server,use -s or config file")
		os.Exit(0)
	}
}

func servers() []sdk.ServerEntity {
	if filter == "" {
		return serverList
	}
	var dest []sdk.ServerEntity
	for _, s := range serverList {
		if s.Addr == filter {
			dest = append(dest, s)
			continue
		}
		for _, tag := range s.Tag {
			if tag == filter {
				dest = append(dest, s)
			}
		}
	}
	return dest
}

func exec_servers(f func(dest sdk.ServerEntity)) {
	for _, s := range servers() {
		fmt.Printf("%s(tag:%v):\n", s.Addr, strings.Join(s.Tag, ","))
		f(s)
	}
}
