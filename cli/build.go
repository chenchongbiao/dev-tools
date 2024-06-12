package cli

import (
	"log"

	"github.com/chenchongbiao/common"
	"github.com/chenchongbiao/core/image"
	"github.com/chenchongbiao/core/rootfs"
	"github.com/chenchongbiao/ios"
	"github.com/rivo/tview"
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
				log.Fatalln("no args, please choose target: [rootfs, board, Docker, WSL]")
			}
			opts.Target = args[0]
			return RunBuild(&opts, nil, nil)
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

func RunBuild(opts *common.BuildOptions, app *tview.Application, textView *tview.TextView) error {
	outCh, errCh := rootfs.CreateRootfsCache(opts)
	ios.CommandOutput(outCh, errCh, app, textView)

	rootfs.CreateRootfsTarFile(opts.DistroName, opts.DistroVersion, opts.Arch)

	if opts.Target == "board" {
		if opts.Device == "" {
			log.Fatalf("not set device, such as: -d qemu")
		}

		if opts.ImageSize == "" {
			opts.ImageSize = "0"
		}

		image.CreateImage(opts)
		// chroot.MountChroot()
		// chroot.UnMountChroot()
	}
	return nil
}
