package chroot

import (
	"fmt"
	"os/exec"
	"time"

	"github.com/chenchongbiao/dev-tools/ios"
	"github.com/chenchongbiao/dev-tools/tools"
	"github.com/rivo/tview"
)

var (
	textView *tview.TextView
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
				tools.PrintLog("No matches mounts; exiting loop.", nil, nil, nil)
				break // 没有找到匹配项，跳出循环
			}
		}

		tools.PrintLog("Match found; continuing operations...", nil, nil, nil)
		ios.Run(fmt.Sprintf(`umount --recursive %s/dev || true`, tools.TmpMountPath()))
		ios.Run(fmt.Sprintf(`umount %s/proc || true`, tools.TmpMountPath()))
		ios.Run(fmt.Sprintf(`umount %s/sys || true`, tools.TmpMountPath()))
		ios.Run(fmt.Sprintf(`umount %s/tmp || true`, tools.TmpMountPath()))
		ios.Run(fmt.Sprintf(`umount %s/var/tmp || true`, tools.TmpMountPath()))
		// -l 懒卸载，避免有程序使用 ROOTFS 还没退出
		ios.Run(fmt.Sprintf("umount -l %s", tools.TmpMountPath()))
		// 添加延时，避免循环过快导致系统资源紧张
		time.Sleep(1 * time.Second)
	}
}

// 在 rootfs 中执行命令
func RunCommandByChoot(rootfsPath, cmd string) {
	tools.PrintLog(fmt.Sprintf("---CMD: %s", fmt.Sprintf("chroot %s /usr/bin/env bash -e -o pipefail -c \"%s\"", rootfsPath, cmd)), nil, nil, textView)
	outCh, errCh := ios.CommandExecutor(fmt.Sprintf("chroot %s /usr/bin/env bash -e -o pipefail -c \"%s\"", rootfsPath, cmd))
	tools.PrintLog("", outCh, errCh, textView)
}
