package chroot

import (
	"fmt"
	"log"
	"os/exec"
	"time"

	"github.com/chenchongbiao/ios"
	"github.com/chenchongbiao/tools"
)

func MountChroot() {
	ios.Run(fmt.Sprintf(`mount -t tmpfs -o "size=99%%" tmpfs "%s/tmp"`, tools.TmpMountPath()))
	ios.Run(fmt.Sprintf(`mount -t tmpfs -o "size=99%%" tmpfs "%s/var/tmp"`, tools.TmpMountPath()))
	ios.Run(fmt.Sprintf(`mount -t proc chproc %s/proc`, tools.TmpMountPath()))
	ios.Run(fmt.Sprintf(`mount -t sysfs chsys %s/sys`, tools.TmpMountPath()))
	ios.Run(fmt.Sprintf(`mount --bind /dev %s/dev`, tools.TmpMountPath()))
}

func UnMountChroot() {
	for {
		err := ios.Run(fmt.Sprintf(`grep -Eq "%s/(dev|proc|sys|tmp|var/tmp)" /proc/mounts`, tools.TmpMountPath()))
		if err != nil {
			// 检查错误是否是因为没有找到匹配项
			if exitError, ok := err.(*exec.ExitError); ok && exitError.ExitCode() == 1 {
				log.Println("No matches mounts; exiting loop.")
				break // 没有找到匹配项，跳出循环
			}
		}

		log.Println("Match found; continuing operations...")
		ios.Run(fmt.Sprintf(`umount --recursive %s/dev || true`, tools.TmpMountPath()))
		ios.Run(fmt.Sprintf(`umount %s/proc || true`, tools.TmpMountPath()))
		ios.Run(fmt.Sprintf(`umount %s/sys || true`, tools.TmpMountPath()))
		ios.Run(fmt.Sprintf(`umount %s/tmp || true`, tools.TmpMountPath()))
		ios.Run(fmt.Sprintf(`umount %s/var/tmp || true`, tools.TmpMountPath()))
		// 添加延时，避免循环过快导致系统资源紧张
		time.Sleep(1 * time.Second)
	}
}

// 在 rootfs 中执行命令
func RunCommandByChoot(rootfsPath, cmd string) {
	ios.Run(fmt.Sprintf("chroot %s /usr/bin/env bash -e -o pipefail -c \"%s\"", tools.TmpMountPath(), cmd))
}
