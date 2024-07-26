package tools

import (
	"fmt"
	"os"
	"path"

	"github.com/chenchongbiao/dev-tools/ios"
)

var (
	sudoUserName = os.Getenv("SUDO_USER")
	userName     = os.Getenv("USER")
)

// 判断是否为 root 用户或者具有 sudo 权限的用户
func IsRootUser() (string, bool) {
	if sudoUserName != "" {
		return sudoUserName, true
	}

	if userName == "root" {
		sudoUserName = userName
		return userName, true
	}

	return userName, false
}

// 修改文件权限
func ModifyFileOwner(file string, tree bool) {
	cmd := "chown"
	if tree {
		cmd += " -R"
	}
	// 修改文件所有者为运行程序的用户
	ios.Run(fmt.Sprintf("%s %s:%s %s", cmd, sudoUserName, sudoUserName, file))
}

func GetUserHome() string {
	return path.Join("/home", sudoUserName)
}
