package rootfs

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/chenchongbiao/common"
	"github.com/chenchongbiao/ios"
	"github.com/chenchongbiao/tools"
)

// 创建新的 rootfs 缓存，并使用 tar 压缩
func CreateRootfsCache(opts *common.BuildOptions) (<-chan string, <-chan string) {
	rootfsPath := GetRootfsPath(opts.DistroName, opts.DistroVersion, opts.Arch)
	if _, err := os.Stat(rootfsPath); err == nil {
		log.Printf("%s is exists", rootfsPath)
		return nil, nil
	}

	// 开启 binfmt 异架构支持
	ios.Run("systemctl start systemd-binfmt")

	args := []string{
		"mmdebstrap",
		"--variant=minbase",
	}
	args = append(args, fmt.Sprintf("--components=%s", opts.Components))

	if opts.Packages == "" {
		opts.Packages = "apt,apt-utils,sudo,vim,bash,bash-completion,ca-certificates,deepin-keyring,parted,network-manager,systemd,systemd-timesyncd,systemd-resolved,curl,screen,vim,init,ssh,kmod,udev,iputils-ping,polkitd,dbus-daemon,grub-efi-arm64,initramfs-tools,uuid-runtime,dmidecode"
	}
	packages := fmt.Sprintf("\"%s\"", opts.Packages)
	args = append(args, fmt.Sprintf("--include=%s", packages))
	args = append(args, fmt.Sprintf("--architectures=%s", opts.Arch))
	args = append(args, opts.DistroVersion)
	args = append(args, rootfsPath)

	sources := fmt.Sprintf("\"%s\"", strings.Replace(opts.Sources, ",", " ", -1))
	args = append(args, sources)

	cmd := strings.Join(args, " ")
	return ios.CommandExecutor(cmd)
}

// 创建 rootfs 的 tar 文件
func CreateRootfsTarFile(distroName, distroVersion, arch string) {
	tarFileName := GetTarFileName(GetRootfsName(distroName, distroVersion, arch))
	rootfsPath := GetRootfsPath(distroName, distroVersion, arch)

	tarFilePath := GetTarFilePath(tarFileName)
	if _, err := os.Stat(tarFilePath); err == nil {
		log.Printf("%s is exists", tarFilePath)
		return
	}

	log.Printf("create %s", tarFileName)

	ios.Run(fmt.Sprintf(`cd %s && tar zfcp %s --xattrs  --exclude='./dev/*' --exclude='./proc/*' \
	--exclude='./run/*' --exclude='./tmp/*' --exclude='./sys/*' --exclude='./home/*' --exclude='./root/*' -C %s .`,
		tools.RootfsCachePath(), tarFileName, rootfsPath))
	tools.ModifyFileOwner(tarFilePath, false)
}

// 解压 rootfs
func ExtractRootfs(rootfsName string) {
	// 解压前清除，目录下的内容
	ios.Run(fmt.Sprintf("rm -rf %s/*", tools.TmpMountPath()))
	tarFileName := GetTarFileName(rootfsName)
	tarFilePath := GetTarFilePath(tarFileName)
	// ios.Run(fmt.Sprintf("tar zxpf %s --xattrs -C %s", tarFilePath, tools.TmpMountPath()))
	exec.Command("tar", "zxpf", tarFilePath, "--xattrs", "-C", tools.TmpMountPath()).Run()
}

func GetRootfsName(distroName, distroVersion, arch string) string {
	return fmt.Sprintf("%s-%s-%s", distroName, distroVersion, arch)
}

func GetRootfsPath(distroName, distroVersion, arch string) string {
	return fmt.Sprintf("%s/%s", tools.RootfsCachePath(), GetRootfsName(distroName, distroVersion, arch))
}

func GetTarFileName(rootfsName string) string {
	return fmt.Sprintf("%s.tar.gz", rootfsName)
}

func GetTarFilePath(tarFileName string) string {
	return fmt.Sprintf("%s/%s", tools.RootfsCachePath(), tarFileName)
}
