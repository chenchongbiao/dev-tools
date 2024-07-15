package cli

import (
	"github.com/chenchongbiao/dev-tools/common"
	"github.com/chenchongbiao/dev-tools/core/run"
	"github.com/chenchongbiao/dev-tools/tools"
	"github.com/spf13/cobra"
)

func BuildCMD() *cobra.Command {
	var opts common.BuildOptions

	cmd := &cobra.Command{
		Use:   "build <target> [flags]",
		Short: "build target",
		Long:  `build rootfs、WSL、board.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				tools.FatalLog("no args, please choose target: [rootfs, board, Docker, WSL]", nil, nil, nil)
			}
			opts.Target = args[0]
			return run.RunBuild(&opts, nil)
		},
	}

	cmd.Flags().StringVarP(&opts.DistroName, "distro-name", "n", "", "Linux distribution")
	cmd.Flags().StringVarP(&opts.DistroVersion, "distro-version", "v", "", "Distribution version")
	cmd.Flags().StringVarP(&opts.Components, "compoents", "c", "", "Distribution version")
	cmd.Flags().StringVarP(&opts.Arch, "arch", "a", "", "System Architecture")
	cmd.Flags().StringVarP(&opts.Sources, "sources", "s", "", "Apt sources")
	cmd.Flags().StringVarP(&opts.Packages, "packages", "p", "", "include package")
	cmd.Flags().StringVarP(&opts.Device, "device", "d", "", "device type")
	cmd.Flags().StringVarP(&opts.ImageSize, "image-size", "i", "", "fixed image size")

	return cmd
}
