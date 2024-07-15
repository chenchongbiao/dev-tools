package run

import (
	"fmt"
	"log"
	"os"

	"github.com/rivo/tview"

	"github.com/chenchongbiao/dev-tools/common"
	"github.com/chenchongbiao/dev-tools/core/chroot"
	"github.com/chenchongbiao/dev-tools/core/image"
	"github.com/chenchongbiao/dev-tools/core/rootfs"
	"github.com/chenchongbiao/dev-tools/tools"
)

func RunBuild(opts *common.BuildOptions, textView *tview.TextView) error {
	outCh, errCh := rootfs.CreateRootfsCache(opts)
	rootfsPath := rootfs.GetRootfsPath(opts.DistroName, opts.DistroVersion, opts.Arch, opts.BaseType)
	tools.PrintLog("", outCh, errCh, textView)
	if _, err := os.Stat(rootfsPath); err != nil {
		tools.PrintLog(fmt.Sprintf("create %s error", rootfsPath), nil, nil, textView)
		return err
	}

	rootfs.CreateRootfsTarFile(opts)

	if opts.Target == "board" {
		if opts.Device == "" {
			log.Fatalf("not set device, such as: -d qemu")
		}

		if opts.ImageSize == "" {
			opts.ImageSize = "0"
		}

		image.CreateImage(opts)
		chroot.MountChroot()
		chroot.UnMountChroot()
	}
	return nil
}
