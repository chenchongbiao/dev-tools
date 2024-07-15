package run

import (
	"log"

	"github.com/rivo/tview"

	"github.com/chenchongbiao/dev-tools/common"
	"github.com/chenchongbiao/dev-tools/core/chroot"
	"github.com/chenchongbiao/dev-tools/core/image"
	"github.com/chenchongbiao/dev-tools/core/rootfs"
	"github.com/chenchongbiao/dev-tools/tools"
)

func RunBuild(opts *common.BuildOptions, textView *tview.TextView) error {
	outCh, errCh := rootfs.CreateRootfsCache(opts)
	tools.PrintLog("", outCh, errCh, textView)

	rootfs.CreateRootfsTarFile(opts.DistroName, opts.DistroVersion, opts.Arch)

	if opts.Target == "board" {
		if opts.Device == "" {
			log.Fatalf("not set device, such as: -d qemu")
		}

		if opts.ImageSize == "" {
			opts.ImageSize = "0"
		}

		image.CreateImage(opts, textView)
		chroot.MountChroot()
		chroot.UnMountChroot()
	}
	return nil
}
