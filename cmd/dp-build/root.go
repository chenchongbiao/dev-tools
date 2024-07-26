/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/chenchongbiao/dev-tools/cli"
	"github.com/chenchongbiao/dev-tools/core/layout"
	"github.com/chenchongbiao/dev-tools/tools"

	"github.com/AlecAivazis/survey/v2/terminal"
	"github.com/spf13/cobra"
)

var (
	version string
)

type exitCode int

const (
	exitOK     exitCode = 0
	exitError  exitCode = 1
	exitCancel exitCode = 2
)

func main() {
	var stderr io.Writer

	hasDebug := os.Getenv("DEBUG") != ""

	RootCmd := execute()
	if cmd, err := RootCmd.ExecuteC(); err != nil {
		if err == tools.SilentError {
			os.Exit(int(exitError))
		} else if tools.IsUserCancellation(err) {
			if errors.Is(err, terminal.InterruptErr) {
				fmt.Fprint(stderr, "\n")
			}

			os.Exit(int(exitCancel))
		}
		tools.PrintError(stderr, err, cmd, hasDebug)
	}
	os.Exit(int(exitOK))
}

func execute() *cobra.Command {
	versionCmd := &cobra.Command{
		Use:     "version",
		Aliases: []string{"ver"},
		Short:   "Print the version of your resto binary.",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("resto version " + version)
		},
	}

	// 判断是否使用 sudo 权限或者 root 用户
	if _, status := tools.IsRootUser(); !status {
		log.Fatalln("Run dp-build requires root or sudo")
		return nil
	}
	tools.CheckDpBuildDot()

	var rootCmd = &cobra.Command{
		Use:   "dp-build <subcommand> [flags]",
		Short: "dp-builder is used to create a universal root filesystem and img image files",
		// 不使用错误处理的默认行为
		RunE: func(cmd *cobra.Command, args []string) error {
			layout.DpBuildLayout()
			return nil
		},
	}

	rootCmd.CompletionOptions.DisableDefaultCmd = true

	rootCmd.AddCommand(
		cli.BuildCMD(),
		versionCmd,
	)
	return rootCmd
}
