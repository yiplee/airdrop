/*
Copyright © 2020 NAME HERE <EMAIL ADDRESS>

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
	"context"

	"github.com/drone/signal"
	propertystore "github.com/fox-one/pkg/store/property"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/yiplee/airdrop/engine"
	taskstore "github.com/yiplee/airdrop/store/task"
)

// engineCmd represents the engine command
var engineCmd = &cobra.Command{
	Use:   "engine",
	Short: "run airdrop engine",
	Run: func(cmd *cobra.Command, args []string) {
		ctx := signal.WithContextFunc(context.Background(), func() {
			logrus.Infoln("interrupt received, terminating process")
		})

		database := provideDatabase()
		broker := provideBroker()
		tasks := taskstore.New(database)
		properties := propertystore.New(database)

		e := engine.Engine{
			Broker:     broker,
			UserID:     cfg.Wallet.UserID,
			Pin:        cfg.Wallet.Pin,
			Tasks:      tasks,
			Properties: properties,
		}

		e.Run(ctx)
	},
}

func init() {
	rootCmd.AddCommand(engineCmd)
}
