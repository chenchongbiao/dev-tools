package common

import "github.com/rivo/tview"

type BuildOptions struct {
	Target        string // 构建目标，rootfs, Docker, WSL,board
	Device        string // 如果是 board 类型，需要指定设备
	DistroName    string // 发行版名称：deepin
	DistroVersion string // 发行版版本：beige
	Components    string // 软件源组件
	Arch          string // 架构
	Sources       string // apt 源
	Packages      string // 需要安装的包
	ImageSize     string // 镜像大小
	BaseType      string // 根文件系统类型， minimal, desktop
	TextView      *tview.TextView
}
