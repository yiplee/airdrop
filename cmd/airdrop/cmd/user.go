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

	"github.com/fox-one/pkg/number"
	"github.com/spf13/cobra"
)

// userCmd represents the user command
var userCmd = &cobra.Command{
	Use:   "user",
	Short: "create a new broker user",
	Run: func(cmd *cobra.Command, args []string) {
		broker := provideBroker()
		name, _ := cmd.Flags().GetString("name")
		pin := number.RandomPin()
		user, err := broker.CreateUser(context.Background(), name, pin)
		if err != nil {
			cmd.PrintErr(err)
			return
		}

		fmt.Printf("%s %s\n", user.UserID, pin)
	},
}

func init() {
	rootCmd.AddCommand(userCmd)
	userCmd.Flags().String("name", "airdrop", "airdrop user name")
}
