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
	"fmt"
	"net/http"
	"time"

	"github.com/drone/signal"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/yiplee/airdrop/handler"
	taskstore "github.com/yiplee/airdrop/store/task"
)

// serverCmd represents the server command
var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "run airdrop server",
	Run: func(cmd *cobra.Command, args []string) {
		database := provideDatabase()
		tasks := taskstore.New(database)

		port, _ := cmd.Flags().GetInt("port")
		addr := fmt.Sprintf(":%d", port)

		srv := &http.Server{
			Addr: addr,
			Handler: handler.Route(
				tasks,
				handler.Config{
					TargetLimit: cfg.Task.MaxTargets,
					BrokerID:    cfg.Wallet.UserID,
					Debug:       enableDebug,
				},
			),
		}

		ctx := context.Background()
		done := make(chan struct{}, 1)
		signal.WithContextFunc(ctx, func() {
			logrus.Debug("shutdown server...")

			// create context with timeout
			ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
			defer cancel()

			if err := srv.Shutdown(ctx); err != nil {
				logrus.WithError(err).Error("graceful shutdown server failed")
			}

			close(done)
		})

		logrus.Println("serve at", addr)
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			logrus.WithError(err).Fatal("server aborted")
		}

		<-done
	},
}

func init() {
	rootCmd.AddCommand(serverCmd)
	serverCmd.Flags().Int("port", 9090, "airdrop server port")
}
