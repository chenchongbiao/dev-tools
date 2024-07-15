package rootfs

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/chenchongbiao/dev-tools/common"
	"github.com/chenchongbiao/dev-tools/ios"
	"github.com/chenchongbiao/dev-tools/tools"
)

// 创建新的 rootfs 缓存，并使用 tar 压缩
func CreateRootfsCache(opts *common.BuildOptions) (<-chan string, <-chan string) {
	rootfsPath := GetRootfsPath(opts.DistroName, opts.DistroVersion, opts.Arch, opts.BaseType)
	if _, err := os.Stat(rootfsPath); err == nil {
		tools.PrintLog(fmt.Sprintf("%s is exists", rootfsPath), nil, nil, nil)
		return nil, nil
	}

	// 开启 binfmt 异架构支持
	ios.Run("systemctl start systemd-binfmt")

	args := []string{
		"mmdebstrap",
		"--variant=minbase",
	}
	args = append(args, fmt.Sprintf("--components=%s", opts.Components))

	if opts.Packages != "" {
		packageList, _ := GetPackageList(opts.BaseType, opts.Arch, opts.Device)
		opts.Packages = fmt.Sprintf("%s,%s", opts.Packages, packageList)
	} else {
		opts.Packages, _ = GetPackageList(opts.BaseType, opts.Arch, opts.Device)
	}

	args = append(args, fmt.Sprintf("--include=%s", fmt.Sprintf("\"%s\"", opts.Packages)))
	args = append(args, fmt.Sprintf("--architectures=%s", opts.Arch))
	args = append(args, opts.DistroVersion)
	args = append(args, rootfsPath)

	sources := fmt.Sprintf("\"%s\"", strings.Replace(opts.Sources, ",", " ", -1))
	args = append(args, sources)

	tools.PrintLog("create rootfs", nil, nil, opts.TextView)
	cmd := strings.Join(args, " ")
	return ios.CommandExecutor(cmd)
}

// 创建 rootfs 的 tar 文件
func CreateRootfsTarFile(opts *common.BuildOptions) {
	tarFileName := GetTarFileName(GetRootfsName(opts.DistroName, opts.DistroVersion, opts.Arch, opts.BaseType))
	rootfsPath := GetRootfsPath(opts.DistroName, opts.DistroVersion, opts.Arch, opts.BaseType)

	tarFilePath := GetTarFilePath(tarFileName)
	if _, err := os.Stat(tarFilePath); err == nil {
		tools.PrintLog(fmt.Sprintf("%s is exists", tarFilePath), nil, nil, opts.TextView)
	}
	tools.PrintLog(fmt.Sprintf("create %s", tarFileName), nil, nil, nil)

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

func GetRootfsName(distroName, distroVersion, arch, baseType string) string {
	return fmt.Sprintf("%s-%s-%s-%s", distroName, distroVersion, arch, baseType)
}

func GetRootfsPath(distroName, distroVersion, arch, baseType string) string {
	return fmt.Sprintf("%s/%s", tools.RootfsCachePath(), GetRootfsName(distroName, distroVersion, arch, baseType))
}

func GetTarFileName(rootfsName string) string {
	return fmt.Sprintf("%s.tar.gz", rootfsName)
}

func GetTarFilePath(tarFileName string) string {
	return fmt.Sprintf("%s/%s", tools.RootfsCachePath(), tarFileName)
}

// 获取软件列表
func GetPackageList(packageType, arch, device string) (string, error) {
	file, err := os.Open(tools.GetPackageListPath(packageType, arch, device))
	if err != nil {
		return "", err
	}
	defer file.Close()

	var content string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		if content != "" {
			content += ","
		}
		cleanLine := strings.Join(strings.FieldsFunc(scanner.Text(), func(r rune) bool {
			return r == ' ' || r == '\t' || r == '\n'
		}), "")
		content += cleanLine
	}

	if err := scanner.Err(); err != nil {
		return "", err
	}

	return content, nil
}
