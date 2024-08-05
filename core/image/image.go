package image

import (
	"fmt"
	"math"
	"path"
	"strconv"

	"github.com/chenchongbiao/dev-tools/core/chroot"
	"github.com/chenchongbiao/dev-tools/core/common"
	"github.com/chenchongbiao/dev-tools/core/rootfs"
	"github.com/chenchongbiao/dev-tools/ios"
	"github.com/chenchongbiao/dev-tools/tools"
)

func CreateImage(opts *common.BuildOptions) {
	rootfsPath := rootfs.GetRootfsPath(opts.DistroName, opts.DistroVersion, opts.Arch, opts.BaseType)
	rootfsSize := ios.RunCommandOutResult(fmt.Sprintf(`du --apparent-size -sm "%s" | cut -f1`, rootfsPath))
	if rootfsSize == "" {
		tools.FatalLog("Error executing du --apparent-size", nil, nil, nil)
	}
	tools.PrintLog(fmt.Sprintf("Current rootfs size: %s MiB", rootfsSize), nil, nil, opts.TextView)

	fixedImageSizeUint, _ := strconv.ParseUint(opts.ImageSize, 10, 64)
	rootfsSizeUint, _ := strconv.ParseUint(rootfsSize, 10, 64)

	sdSize := fixedImageSizeUint
	// config 分区, > 大于0，则表示需要该分区
	configPart := -1
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

	if opts.Device == "rock-5b" {
		configPart = next
		next++
	}

	uefiPart := next
	next++
	rootPart := next

	tools.PrintLog("create partition table", nil, nil, opts.TextView)

	if configPart > 0 {
		configSize := 16
		ios.Run(fmt.Sprintf(`(echo n; echo %d; echo "32768"; echo +%dM; echo 0700; \
		echo c; echo config; \
		echo n; echo %d; echo "65536"; echo +%dM; echo ef00; \
		echo c; echo %d; echo boot; \
		echo n; echo %d; echo ""; echo ""; echo "ef00"; \
		echo c; echo %d; echo rootfs; \
		echo w; echo y) | gdisk %s`, configPart, configSize, uefiPart, uefiSize, uefiPart, rootPart, rootPart, imagePath))
	} else {
		ios.Run(fmt.Sprintf(`(echo n; echo %d; echo ""; echo +%dM;  echo ef00; \
		echo c; echo efi; \
		echo n; echo %d; echo ""; echo ""; echo ""; \
		echo c; echo %d; echo rootfs; \
		echo w; echo y) | gdisk %s`, uefiPart, uefiSize, rootPart, rootPart, imagePath))
	}

	loop := ios.RunCommandOutResult(fmt.Sprintf(`losetup --partscan --find --show %s`, imagePath))
	tools.PrintLog(fmt.Sprintf("Allocated loop device %s", loop), nil, nil, opts.TextView)

	deviceConfigPath := tools.GetDeviceConfigPath(opts.Arch, opts.Device)

	// 设置这些文件系统的标签。dosfslabel 是用来设置vfat（FAT）文件系统的标签，e2label 是用来设置ext2/ext3/ext4文件系统的标签
	ios.Run(fmt.Sprintf("mkfs.vfat -F32 %sp%d", loop, uefiPart))
	ios.Run(fmt.Sprintf("mmd -i %sp%d ::/EFI", loop, uefiPart))
	// ios.Run(fmt.Sprintf("mmd -i %sp%d ::/EFI/BOOT", loop, uefiPart))
	ios.Run(fmt.Sprintf("mcopy -i %sp%d %s/* ::/EFI", loop, uefiPart, path.Join(deviceConfigPath, "EFI")))
	ios.Run(fmt.Sprintf("e2label %sp%d efi", loop, rootPart))

	ios.Run(fmt.Sprintf("mkfs.ext4 %sp%d", loop, rootPart))
	ios.Run(fmt.Sprintf("e2label %sp%d rootfs", loop, rootPart))

	if configPart > 0 {
		ios.Run(fmt.Sprintf("mkfs.vfat -F32 %sp%d", loop, configPart))
		ios.Run(fmt.Sprintf("e2label %sp%d config", loop, configPart))
		ios.Run(fmt.Sprintf("fatlabel %sp%d config", loop, configPart))
	}

	// 解压之前先做一次卸载目录
	tools.PrintLog("umount chroot", nil, nil, opts.TextView)
	chroot.UnMountChroot()

	// 挂载设备
	ios.Run(fmt.Sprintf("mount %sp%d %s", loop, rootPart, tools.TmpMountPath()))
	rootfsName := rootfs.GetRootfsName(opts.DistroName, opts.DistroVersion, opts.Arch, opts.BaseType)
	rootfs.ExtractRootfs(rootfsName)
	chroot.MountChroot()

	ios.Run(fmt.Sprintf("mkdir %s/boot/efi", tools.TmpMountPath()))
	ios.Run(fmt.Sprintf("mount %sp%d %s/boot/efi", loop, uefiPart, tools.TmpMountPath()))

	tools.PrintLog(fmt.Sprintf("copy grup to %s/boot", tools.TmpMountPath()), nil, nil, opts.TextView)
	ios.Run(fmt.Sprintf("cp -r %s/grub/ %s/boot", deviceConfigPath, tools.TmpMountPath()))

	tools.PrintLog(fmt.Sprintf("copy kernel to %s/boot", tools.TmpMountPath()), nil, nil, opts.TextView)
	ios.Run(fmt.Sprintf("cp -r %s/kernel/* %s/boot", deviceConfigPath, tools.TmpMountPath()))
	ios.Run(fmt.Sprintf("mkdir %s/lib/modules", tools.TmpMountPath()))
	ios.Run(fmt.Sprintf("cp -r %s/modules/* %s/lib/modules", deviceConfigPath, tools.TmpMountPath()))

	tools.PrintLog(fmt.Sprintf("copy sources.list.d to %s/etc/apt/sources.list.d", tools.TmpMountPath()), nil, nil, opts.TextView)
	ios.Run(fmt.Sprintf("cp -r %s/sources.list.d/* %s/etc/apt/sources.list.d", deviceConfigPath, tools.TmpMountPath()))

	if opts.Arch == "arm64" {
		if opts.Device == "qemu" {
			tools.PrintLog(fmt.Sprintf("add fat to %s/etc/modules", tools.TmpMountPath()), nil, nil, opts.TextView)
			ios.Run(fmt.Sprintf("echo \"fat\" >> %s/etc/modules", tools.TmpMountPath()))
			ios.Run(fmt.Sprintf("echo \"vfat\" >> %s/etc/modules", tools.TmpMountPath()))
		}
		if opts.Device == "rock-5b" {
			ios.Run(fmt.Sprintf("mkdir %s/config", tools.TmpMountPath()))

			tools.PrintLog(fmt.Sprintf("copy extra packages to %s/tmp", tools.TmpMountPath()), nil, nil, opts.TextView)
			ios.Run(fmt.Sprintf("cp -r %s/extra-packages/* %s/tmp", deviceConfigPath, tools.TmpMountPath()))
			tools.PrintLog(fmt.Sprintf("installing extra packages from %s/tmp", tools.TmpMountPath()), nil, nil, opts.TextView)
			chroot.RunCommandByChoot(tools.TmpMountPath(), "cd /tmp && ls *deb")
			chroot.RunCommandByChoot(tools.TmpMountPath(), "cd /tmp && apt install -y ./*deb")
			chroot.RunCommandByChoot(tools.TmpMountPath(), `apt update && apt install -y \
			task-rock-5b linux-image-rock-5b linux-headers-rock-5b u-boot-rock-5b \
			radxa-bootutils radxa-firmware radxa-otgutils radxa-udev \
			radxa-system-config-kernel-cmdline-ttyfiq0 \
			rfkill rsetup-config-first-boot \
			u-boot-tools efibootmgr systemd-boot apt-listchanges \
			pipewire-audio avahi-daemon`)

			chroot.RunCommandByChoot(tools.TmpMountPath(), fmt.Sprintf("cd /usr/lib/u-boot/%s && ./setup.sh update_bootloader %s rk3588", opts.Device, loop))
			chroot.RunCommandByChoot(tools.TmpMountPath(), "u-boot-update")

			ios.Run(fmt.Sprintf("mount %sp%d %s/config", loop, configPart, tools.TmpMountPath()))
			tools.PrintLog(fmt.Sprintf("copy config to %s/config", tools.TmpMountPath()), nil, nil, opts.TextView)
			ios.Run(fmt.Sprintf("cp -r %s/config/* %s/config", deviceConfigPath, tools.TmpMountPath()))
		}
	}

	tools.PrintLog("generate /etc/fstab", nil, nil, opts.TextView)
	rootPartUuid := ios.RunCommandOutResult(fmt.Sprintf("blkid -s UUID -o value %sp%d", loop, rootPart))

	tools.PrintLog(fmt.Sprintf("root uuid: %s", rootPartUuid), nil, nil, opts.TextView)
	ios.Run(fmt.Sprintf("echo \"UUID=%s / ext4 defaults 0 1\" > %s/etc/fstab", rootPartUuid, tools.TmpMountPath()))
	if configPart > 0 {
		configPartUuid := ios.RunCommandOutResult(fmt.Sprintf("blkid -s UUID -o value %sp%d", loop, configPart))
		tools.PrintLog(fmt.Sprintf("config uuid: %s", configPartUuid), nil, nil, opts.TextView)
		ios.Run(fmt.Sprintf("echo \"UUID=%s /config vfat defaults,x-systemd.automount 0 2\" >> %s/etc/fstab", configPartUuid, tools.TmpMountPath()))
	}

	uefiPartUuid := ios.RunCommandOutResult(fmt.Sprintf("blkid -s UUID -o value %sp%d", loop, uefiPart))
	tools.PrintLog(fmt.Sprintf("efi uuid: %s", uefiPartUuid), nil, nil, opts.TextView)
	ios.Run(fmt.Sprintf("echo \"UUID=%s /boot/efi vfat defaults,x-systemd.automount 0 2\" >> %s/etc/fstab", uefiPartUuid, tools.TmpMountPath()))

	ios.Run(fmt.Sprintf("sed -i \"s/efi_uuid/%s/g\" %s/boot/grub/grub.cfg", uefiPartUuid, tools.TmpMountPath()))
	ios.Run(fmt.Sprintf("sed -i \"s/root_uuid/%s/g\" %s/boot/grub/grub.cfg", rootPartUuid, tools.TmpMountPath()))

	ios.Run(fmt.Sprintf("sed -i \"s/root_uuid/%s/g\" %s/boot/efi/EFI/boot/grub.cfg", rootPartUuid, tools.TmpMountPath()))

	tools.PrintLog("set hostname", nil, nil, opts.TextView)
	ios.Run(fmt.Sprintf("echo \"deepin-%s-%s\" | tee %s/etc/hostname", opts.Arch, opts.Device, tools.TmpMountPath()))

	tools.PrintLog("update-initramfs -u", nil, nil, opts.TextView)
	chroot.RunCommandByChoot(tools.TmpMountPath(), "update-initramfs -u")

	rootfs.ConfigureUser()

	chroot.UnMountChroot()
	ios.Run(fmt.Sprintf("losetup -D /dev/%s", loop))

	fixImageFile(imagePath)

	tools.ModifyFileOwner(imagePath, false)
}

// 只创建包含根文件系统的镜像，不做分区
func CreateOnlyRootfsImage(opts *common.BuildOptions) {
	rootfsPath := rootfs.GetRootfsPath(opts.DistroName, opts.DistroVersion, opts.Arch, opts.BaseType)
	rootfsSize := ios.RunCommandOutResult(fmt.Sprintf(`du --apparent-size -sm "%s" | cut -f1`, rootfsPath))
	if rootfsSize == "" {
		tools.FatalLog("Error executing du --apparent-size", nil, nil, nil)
	}
	tools.PrintLog(fmt.Sprintf("Current rootfs size: %s MiB", rootfsSize), nil, nil, opts.TextView)

	fixedImageSizeUint, _ := strconv.ParseUint(opts.ImageSize, 10, 64)
	rootfsSizeUint, _ := strconv.ParseUint(rootfsSize, 10, 64)

	extraRootfsSize := 500
	sdSize := fixedImageSizeUint
	if fixedImageSizeUint < rootfsSizeUint {
		fixedImageSizeUint = rootfsSizeUint + uint64(extraRootfsSize)
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

	ios.Run(fmt.Sprintf("mkfs.ext4 -F -m 0 -L rootfs %s", imagePath))

	// 解压之前先做一次卸载目录
	tools.PrintLog("umount chroot", nil, nil, opts.TextView)
	chroot.UnMountChroot()

	loop := ios.RunCommandOutResult(fmt.Sprintf(`losetup --partscan --find --show %s`, imagePath))
	tools.PrintLog(fmt.Sprintf("Allocated loop device %s", loop), nil, nil, opts.TextView)

	// 挂载设备
	ios.Run(fmt.Sprintf("mount %s %s", loop, tools.TmpMountPath()))

	deviceConfigPath := tools.GetDeviceConfigPath(opts.Arch, opts.Device)

	rootfsName := rootfs.GetRootfsName(opts.DistroName, opts.DistroVersion, opts.Arch, opts.BaseType)
	rootfs.ExtractRootfs(rootfsName)
	chroot.MountChroot()

	if opts.Device == "mipad5" {
		tools.PrintLog("generate /etc/fstab", nil, nil, opts.TextView)
		rootfsUuid := ios.RunCommandOutResult(fmt.Sprintf("blkid -s UUID -o value %s", loop))
		tools.PrintLog(fmt.Sprintf("root uuid: %s", rootfsUuid), nil, nil, opts.TextView)
		ios.Run(fmt.Sprintf("echo \"UUID=%s / ext4 defaults 0 1\" > %s/etc/fstab", rootfsUuid, tools.TmpMountPath()))
		ios.Run(fmt.Sprintf("echo \"PARTLABEL=esp /boot/efi vfat umask=0077 0 1\" >> %s/etc/fstab", tools.TmpMountPath()))

		ios.Run(fmt.Sprintf("mkdir %s/boot/efi", tools.TmpMountPath()))
		tools.PrintLog(fmt.Sprintf("copy efi to %s/boot/efi", tools.TmpMountPath()), nil, nil, opts.TextView)
		ios.Run(fmt.Sprintf("cp -r %s/EFI/* %s/boot/efi", deviceConfigPath, tools.TmpMountPath()))

		tools.PrintLog(fmt.Sprintf("copy grup to %s/boot", tools.TmpMountPath()), nil, nil, opts.TextView)
		ios.Run(fmt.Sprintf("cp -r %s/grub/ %s/boot", deviceConfigPath, tools.TmpMountPath()))

		tools.PrintLog(fmt.Sprintf("copy kernel to %s/boot", tools.TmpMountPath()), nil, nil, opts.TextView)
		ios.Run(fmt.Sprintf("cp -r %s/kernel/* %s/boot", deviceConfigPath, tools.TmpMountPath()))
		ios.Run(fmt.Sprintf("mkdir %s/lib/modules", tools.TmpMountPath()))
		ios.Run(fmt.Sprintf("cp -r %s/modules/* %s/lib/modules", deviceConfigPath, tools.TmpMountPath()))

		tools.PrintLog(fmt.Sprintf("copy firmware to %s/lib", tools.TmpMountPath()), nil, nil, opts.TextView)
		ios.Run(fmt.Sprintf("mkdir %s/lib/firmware", tools.TmpMountPath()))
		ios.Run(fmt.Sprintf("cp -r %s/firmware/* %s/lib/firmware", deviceConfigPath, tools.TmpMountPath()))

		tools.PrintLog(fmt.Sprintf("copy ucm2 to %s/usr/share/alsa", tools.TmpMountPath()), nil, nil, opts.TextView)
		ios.Run(fmt.Sprintf("mkdir -p %s/usr/share/alsa/ucm2", tools.TmpMountPath()))
		ios.Run(fmt.Sprintf("cp -r %s/ucm2/* %s/usr/share/alsa/ucm2", deviceConfigPath, tools.TmpMountPath()))
	}

	tools.PrintLog("set hostname", nil, nil, opts.TextView)
	ios.Run(fmt.Sprintf("echo \"deepin-%s-%s\" | tee %s/etc/hostname", opts.Arch, opts.Device, tools.TmpMountPath()))

	rootfs.ConfigureUser()

	tools.PrintLog("update-initramfs -u", nil, nil, opts.TextView)
	chroot.RunCommandByChoot(tools.TmpMountPath(), "update-initramfs -u")

	chroot.UnMountChroot()
	ios.Run(fmt.Sprintf("losetup -D /dev/%s", loop))

	fixImageFile(imagePath)

	tools.ModifyFileOwner(imagePath, false)
}

// 修复镜像文件
func fixImageFile(img string) {
	ios.Run(fmt.Sprintf("e2fsck -p -f %s", img))
	ios.Run(fmt.Sprintf("resize2fs -M %s", img))
}

func GetImageName(distroName, distroVersion, device, arch, baseType string) string {
	return fmt.Sprintf("%s-%s-%s-%s-%s.img", distroName, distroVersion, device, arch, baseType)
}

func GetImagePath(distroName, distroVersion, device, arch, baseType string) string {
	return fmt.Sprintf("%s/%s", tools.OutputImagePath(), GetImageName(distroName, distroVersion, device, arch, baseType))
}
