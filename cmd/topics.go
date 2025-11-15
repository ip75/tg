/*
Copyright © 2025 <admin@goswami.ru>
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// topicsCmd represents the topics command
var topicsCmd = &cobra.Command{
	Use:   "topics",
	Short: "Operations with telegram topics",
	Long:  `Operations with telegram topics: update, list etc.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("topics called")
	},
}

func init() {
	rootCmd.AddCommand(topicsCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// topicsCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// topicsCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
