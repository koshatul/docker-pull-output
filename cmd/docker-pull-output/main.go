package main

import (
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

//nolint:gochecknoglobals // cobra uses globals in main
var rootCmd = &cobra.Command{
	Use:   "docker-pull-ci",
	Short: "Command to parse output of docker pull and docker push for readability on long push/pulls",
	Run:   parseCommand,
	Args:  cobra.NoArgs,
}

//nolint:gochecknoinits // init is used in main for cobra
func init() {
	cobra.OnInitialize(beforeAction)
}

func beforeAction() {
	logrus.SetOutput(os.Stderr)
}

func main() {
	_ = rootCmd.Execute()
}
