/*
Copyright 2024 Kevin Park

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
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/krapie/l7/internal"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "l7",
	Short: "l7 is a layer 7 load balancer built from scratch in Go.",
	Long:  `l7 is a layer 7 load balancer built from scratch in Go.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := runAgent(cmd, args); err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

		os.Exit(0)
	},
}

func runAgent(cmd *cobra.Command, args []string) error {
	serviceDiscoveryMode, err := cmd.Flags().GetString("service-discovery-mode")
	if err != nil {
		return err
	}

	targetFilter, err := cmd.Flags().GetString("target-filter")
	if err != nil {
		return err
	}

	maglevHashKey, err := cmd.Flags().GetString("maglev-hash-key")
	if err != nil {
		return err
	}

	agent, err := internal.NewAgent(&internal.Config{
		ServiceDiscoveryMode: serviceDiscoveryMode,
		TargetFilter:         targetFilter,
		MaglevHashKey:        maglevHashKey,
	})
	if err != nil {
		return err
	}

	if err = agent.Start(); err != nil {
		return err
	}

	if code := handleSignal(agent); code != 0 {
		return fmt.Errorf("exit code: %d", code)
	}

	return nil
}

func handleSignal(agent *internal.Agent) int {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)

	var sig os.Signal
	select {
	case s := <-sigCh:
		sig = s
	case <-agent.ShutdownCh():
		return 0
	}

	graceful := false
	if sig == syscall.SIGINT || sig == syscall.SIGTERM {
		graceful = true
	}

	gracefulCh := make(chan struct{})
	go func() {
		if err := agent.Shutdown(graceful); err != nil {
			return
		}
		close(gracefulCh)
	}()

	gracefulTimeout := 5 * time.Second
	select {
	case <-sigCh:
		return 1
	case <-time.After(gracefulTimeout):
		return 1
	case <-gracefulCh:
		return 0
	}
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.l7.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	rootCmd.Flags().String("service-discovery-mode", "docker", "Service discovery mode")
	rootCmd.Flags().String("target-filter", "traefik/whoami", "Backend target filter for service discovery")
	rootCmd.Flags().String("maglev-hash-key", "X-Shard-Key", "Hash key for maglev consistent hashing")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".l7" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".l7")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}
