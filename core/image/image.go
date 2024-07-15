package image

import (
	"fmt"
	"math"
	"path"
	"strconv"

	"github.com/chenchongbiao/dev-tools/common"
	"github.com/chenchongbiao/dev-tools/core/chroot"
	"github.com/chenchongbiao/dev-tools/core/rootfs"
	"github.com/chenchongbiao/dev-tools/ios"
	"github.com/chenchongbiao/dev-tools/tools"
)

func CreateImage(opts *common.BuildOptions) {
	rootfsPath := rootfs.GetRootfsPath(opts.DistroName, opts.DistroVersion, opts.Arch, opts.BaseType)
	// imagePath := fmt.Sprintf("%s/%s", tools.OutputImagePath(), imageName)
	rootfsSize := ios.RunCommandOutResult(fmt.Sprintf(`du --apparent-size -sm "%s" | cut -f1`, rootfsPath))
	if rootfsSize == "" {
		tools.FatalLog("Error executing du --apparent-size", nil, nil, nil)
	}
	tools.PrintLog(fmt.Sprintf("Current rootfs size: %s MiB", rootfsSize), nil, nil, opts.TextView)

	fixedImageSizeUint, _ := strconv.ParseUint(opts.ImageSize, 10, 64)
	rootfsSizeUint, _ := strconv.ParseUint(rootfsSize, 10, 64)

	sdSize := fixedImageSizeUint
	// default efi size
	uefiSize := 300
	extraRootfsSize := 500
	// UefiMountPoint := "/boot/efi"

	if fixedImageSizeUint < rootfsSizeUint {
		fixedImageSizeUint = rootfsSizeUint + uint64(uefiSize) + uint64(extraRootfsSize)
		// 计算最终需要生成的镜像大小对齐至4MiB
		// 再进行扩展，生成镜像的大小
		if opts.BaseType == "minimal" {
			sdSize = uint64(math.Ceil(float64(fixedImageSizeUint)*1.3/4) * 4)
		} else {
			// 安装桌面, 需要更大的空间
			sdSize = uint64(math.Ceil(float64(fixedImageSizeUint)*2/4) * 4)
		}
	}

	imagePath := GetImagePath(opts.DistroName, opts.DistroVersion, opts.Device, opts.Arch, opts.BaseType)

	tools.PrintLog(fmt.Sprintf("Creating image: %s, sdsize %d MiB", imagePath, sdSize), nil, nil, opts.TextView)
	ios.Run(fmt.Sprintf(`dd if=/dev/zero of=%s bs=1M count=%d`, imagePath, sdSize))

	// 分区顺序
	next := 1
	uefiPart := next
	next++
	rootPart := next

	if opts.Arch == "arm64" && opts.Device == "qemu" {
		uefiPart = 15
		rootPart = 1
	}

	tools.PrintLog("create partition table", nil, nil, opts.TextView)

	ios.Run(fmt.Sprintf(`(echo n; echo %d; echo ""; echo +%dM;  echo ef00; echo n; echo %d; echo ""; echo ""; echo ""; echo w; echo y) | gdisk %s`, uefiPart, uefiSize, rootPart, imagePath))

	loop := ios.RunCommandOutResult(fmt.Sprintf(`losetup --partscan --find --show %s`, imagePath))
	tools.PrintLog(fmt.Sprintf("Allocated loop device %s", loop), nil, nil, opts.TextView)

	deviceConfigPath := tools.GetDeviceConfigPath(opts.Arch, opts.Device)
	// 设置这些文件系统的标签。dosfslabel 是用来设置vfat（FAT）文件系统的标签，e2label 是用来设置ext2/ext3/ext4文件系统的标签
	ios.Run(fmt.Sprintf("mkfs.vfat %sp%d", loop, uefiPart))
	ios.Run(fmt.Sprintf("mmd -i %sp%d ::/EFI", loop, uefiPart))
	ios.Run(fmt.Sprintf("mmd -i %sp%d ::/EFI/BOOT", loop, uefiPart))
	ios.Run(fmt.Sprintf("mcopy -i  %sp%d %s/* ::/EFI/BOOT", loop, uefiPart, path.Join(deviceConfigPath, "EFI")))

	ios.Run(fmt.Sprintf("mkfs.ext4 %sp%d", loop, rootPart))
	ios.Run(fmt.Sprintf("e2label %sp%d", loop, rootPart))

	// 解压之前先做一次卸载目录
	tools.PrintLog("umount chroot", nil, nil, opts.TextView)
	chroot.UnMountChroot()

	// 挂载设备
	ios.Run(fmt.Sprintf("mount %sp%d %s", loop, rootPart, tools.TmpMountPath()))

	rootfsName := rootfs.GetRootfsName(opts.DistroName, opts.DistroVersion, opts.Arch, opts.BaseType)
	rootfs.ExtractRootfs(rootfsName)
	chroot.MountChroot()

	ios.Run(fmt.Sprintf("mkdir %s/boot/efi", tools.TmpMountPath()))

	tools.PrintLog(fmt.Sprintf("copy grup to %s/boot", tools.TmpMountPath()), nil, nil, opts.TextView)
	ios.Run(fmt.Sprintf("cp -r %s/grub/ %s/boot", deviceConfigPath, tools.TmpMountPath()))

	tools.PrintLog(fmt.Sprintf("copy kernel to %s/boot", tools.TmpMountPath()), nil, nil, opts.TextView)
	tools.PrintLog(fmt.Sprintf("copy kernel to %s/boot", tools.TmpMountPath()), nil, nil, opts.TextView)
	ios.Run(fmt.Sprintf("cp -r %s/kernel/* %s/boot", deviceConfigPath, tools.TmpMountPath()))
	ios.Run(fmt.Sprintf("mkdir %s/lib/modules", tools.TmpMountPath()))
	ios.Run(fmt.Sprintf("cp -r %s/modules/* %s/lib/modules", deviceConfigPath, tools.TmpMountPath()))

	if opts.Arch == "arm64" && opts.Device == "qemu" {
		tools.PrintLog(fmt.Sprintf("copy kernel to %s/etc/modules", tools.TmpMountPath()), nil, nil, opts.TextView)
		ios.Run(fmt.Sprintf("echo \"fat\" >> %s/etc/modules", tools.TmpMountPath()))
		ios.Run(fmt.Sprintf("echo \"vfat\" >> %s/etc/modules", tools.TmpMountPath()))
	}

	tools.PrintLog("generate /etc/fstab", nil, nil, opts.TextView)
	rootPartUuid := ios.RunCommandOutResult(fmt.Sprintf("blkid -s UUID -o value %sp%d", loop, rootPart))
	tools.PrintLog(fmt.Sprintf("root uuid: %s", rootPartUuid), nil, nil, opts.TextView)
	ios.Run(fmt.Sprintf("echo \"UUID=%s / ext4 rw,discard,errors=remount-ro,x-systemd.growfs 0 1\" >> %s/etc/fstab", rootPartUuid, tools.TmpMountPath()))

	uefiPartUuid := ios.RunCommandOutResult(fmt.Sprintf("blkid -s UUID -o value %sp%d", loop, uefiPart))
	tools.PrintLog(fmt.Sprintf("efi uuid: %s", uefiPartUuid), nil, nil, opts.TextView)
	ios.Run(fmt.Sprintf("echo \"UUID=%s /boot/efi vfat defaults 0 2\" >> %s/etc/fstab", uefiPartUuid, tools.TmpMountPath()))

	ios.Run(fmt.Sprintf("sed -i \"s/root_uuid/%s/g\" %s/boot/grub/grub.cfg", rootPartUuid, tools.TmpMountPath()))
	tools.PrintLog("set hostname", nil, nil, opts.TextView)
	ios.Run(fmt.Sprintf("echo \"deepin-%s-%s\" | tee %s/etc/hostname", opts.Arch, opts.Device, tools.TmpMountPath()))

	chroot.RunCommandByChoot(rootfsPath, "useradd  -s /bin/bash -m -g users deepin")
	chroot.RunCommandByChoot(rootfsPath, "usermod -a -G sudo deepin")
	chroot.RunCommandByChoot(rootfsPath, "chsh -s -a -G sudo deepin")
	chroot.RunCommandByChoot(rootfsPath, "echo root:deepin | chpasswd")
	chroot.RunCommandByChoot(rootfsPath, "echo deepin:deepin | chpasswd")

	chroot.UnMountChroot()
	// # -l 懒卸载，避免有程序使用 ROOTFS 还没退出
	ios.Run(fmt.Sprintf("umount -l %s", tools.TmpMountPath()))
	ios.Run(fmt.Sprintf("losetup -D /dev/%s", loop))

	tools.ModifyFileOwner(imagePath, false)
}

func GetImageName(distroName, distroVersion, device, arch, baseType string) string {
	return fmt.Sprintf("%s-%s-%s-%s-%s.img", distroName, distroVersion, device, arch, baseType)
}

func GetImagePath(distroName, distroVersion, device, arch, baseType string) string {
	return fmt.Sprintf("%s/%s", tools.OutputImagePath(), GetImageName(distroName, distroVersion, device, arch, baseType))
}
