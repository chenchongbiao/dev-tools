package tools

import (
	"os"

	"github.com/tidwall/pretty"
)

func SettingsContent() string {
	stgFile, err := os.ReadFile(dpBuildSettingsFile)

	if err != nil {
		panic(err)
	}

	return string(stgFile)
}

func SetDefaultSettings() {
	defaultSettings := `
		{
			"rootfs": {
				"minimal": "apt,apt-utils,bash-completion,curl,sudo,vim,bash,ca-certificates,deepin-keyring,init,ssh,net-tools,iputils-ping,lshw,iproute2,iptables,locales,procps",
				"desktop": ""
			},
			"board": {
				"minimal": "dmidecode,adduser,uuid-runtime,iw,initramfs-tools,polkitd,dbus-daemon,network-manager,systemd,systemd-timesyncd,systemd-resolved,kmod,udev,parted,pciutils,ldnsutils",
				"desktop": ""
			}
		}
	`
	prettySettings := pretty.Pretty([]byte(defaultSettings))

	err := os.WriteFile(dpBuildSettingsFile, []byte(string(prettySettings)), 0644)

	if err != nil {
		panic(err)
	}
}
