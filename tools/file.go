package tools

import (
	"os"
	"path"
)

var (
	pwd, _              = os.Getwd()
	dpBuildDot          = path.Join(GetUserHome(), ".dp-build")
	dpBuildSettingsFile = path.Join(dpBuildDot, "settings.json")
	cachePath           = path.Join(pwd, "cache")
	rootfsCachePath     = path.Join(cachePath, "rootfs")
	tmpMountPath        = path.Join(cachePath, "tmp")
	outputPath          = path.Join(pwd, "output")
	outputImagePath     = path.Join(outputPath, "image")
)

// 用来检查并创建 dp-build 需要的目录或者文件
func CheckDpBuildDot() {
	if _, err := os.Stat(dpBuildDot); os.IsNotExist(err) {
		os.Mkdir(dpBuildDot, 0755)
		ModifyFileOwner(dpBuildDot, false)
	}

	if _, err := os.Stat(dpBuildSettingsFile); os.IsNotExist(err) {
		os.Create(dpBuildSettingsFile)
		ModifyFileOwner(dpBuildSettingsFile, false)
		SetDefaultSettings()
	}

	if _, err := os.Stat(cachePath); os.IsNotExist(err) {
		os.MkdirAll(cachePath, 0755)
		ModifyFileOwner(cachePath, true)
	}

	if _, err := os.Stat(rootfsCachePath); os.IsNotExist(err) {
		os.Mkdir(rootfsCachePath, 0755)
		ModifyFileOwner(rootfsCachePath, true)
	}

	if _, err := os.Stat(tmpMountPath); os.IsNotExist(err) {
		os.Mkdir(tmpMountPath, 0755)
		ModifyFileOwner(tmpMountPath, false)
	}

	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		os.MkdirAll(outputPath, 0755)
		os.Mkdir(outputImagePath, 0755)
		ModifyFileOwner(tmpMountPath, true)
	}
}

// 返回 dp-build 的 settings 文件
func DpBuildSettingsFile() string {
	return dpBuildSettingsFile
}

// 返回 rootfs 的缓存路径
func RootfsCachePath() string {
	return rootfsCachePath
}

// 临时目录，用来挂载 chroot
func TmpMountPath() string {
	return tmpMountPath
}

// 生成磁盘镜像的路径
func OutputImagePath() string {
	return outputImagePath
}

// 返回设备存放配置的路径
func GetDeviceConfigPath(arch, device string) string {
	return path.Join(pwd, "config", arch, device)
}

// 获取预装包列表的路径
func GetPackageListPath(packageType, arch, device string) string {
	return path.Join(GetDeviceConfigPath(arch, device), path.Join("package", packageType))
}
