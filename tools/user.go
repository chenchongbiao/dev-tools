package tools

import (
	"fmt"
	"os"

	"github.com/chenchongbiao/dev-tools/ios"
)

var (
	sudoUser = os.Getenv("SUDO_USER")
	user     = os.Getenv("USER")
)

// 判断是否为 root 用户或者具有 sudo 权限的用户
func IsRootUser() (string, bool) {
	if sudoUser != "" {
		return sudoUser, true
	}

	if user == "root" {
		return user, true
	}

	return user, false
}

// 修改文件权限
func ModifyFileOwner(file string, tree bool) {
	cmd := "chown"
	if tree {
		cmd += " -R"
	}
	// 修改文件所有者为运行程序的用户
	ios.Run(fmt.Sprintf("%s %s:%s %s", cmd, sudoUser, sudoUser, file))
}
