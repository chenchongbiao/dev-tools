package run

import (
	"fmt"
	"log"
	"os"

	"github.com/rivo/tview"

	"github.com/chenchongbiao/dev-tools/core/chroot"
	"github.com/chenchongbiao/dev-tools/core/common"
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

	tarFileName := rootfs.GetTarFileName(rootfs.GetRootfsName(opts.DistroName, opts.DistroVersion, opts.Arch, opts.BaseType))
	tarFilePath := rootfs.GetTarFilePath(tarFileName)
	if _, err := os.Stat(tarFilePath); err != nil {
		tools.PrintLog("locale-gen", nil, nil, opts.TextView)
		chroot.RunCommandByChoot(rootfsPath, `sed -i -E 's/#[[:space:]]?(zh_CN.UTF-8[[:space:]]+UTF-8)/\1/g' /etc/locale.gen`)
		chroot.RunCommandByChoot(rootfsPath, `sed -i -E 's/#[[:space:]]?(en_US.UTF-8[[:space:]]+UTF-8)/\1/g' /etc/locale.gen`)
		chroot.RunCommandByChoot(rootfsPath, "locale-gen")
		chroot.RunCommandByChoot(rootfsPath, "DEBIAN_FRONTEND=noninteractive dpkg-reconfigure locales")
		tools.PrintLog("Annotation USERS_GID and USERS_GROUP", nil, nil, opts.TextView)
		// 微软提供的 wsl 启动器会调用adduser,需要将 USERS_GID 和 USERS_GROUP 注释。
		chroot.RunCommandByChoot(rootfsPath, `sed -i -e 's/USERS_GID=100/#USERS_GID=100/' -e 's/USERS_GROUP=users/#USERS_GROUP=users/' /etc/adduser.conf`)
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
	}
	return nil
}
