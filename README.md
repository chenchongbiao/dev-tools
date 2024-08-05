# 介绍

# 构建

# 使用

## 构建根文件系统

```bash
sudo ./bin/dp-build build rootfs -n="deepin" -v="beige" -c="main,commercial,community" -a="arm64" -s="deb https://community-packages.deepin.com/beige/ beige main commercial community"
```

## 构建根文件系统的镜像

### 小米平板5

```bash
sudo ./bin/dp-build build rootfsimg -n="deepin" -v="beige" -c="main,commercial,community" -a="arm64" -s="deb https://community-packages.deepin.com/beige/ beige main commercial community" -d "mipad5"
```

## 构建开发板

### qemu

```bash
sudo ./bin/dp-build build board -n="deepin" -v="beige" -c="main,commercial,community" -a="arm64" -s="deb https://community-packages.deepin.com/beige/ beige main commercial community" -d qemu
```

### rock-5b

```bash
sudo ./bin/dp-build build board -n="deepin" -v="beige" -c="main,commercial,community" -a="arm64" -s="deb https://community-packages.deepin.com/beige/ beige main commercial community" -d rock-5b
```

## 指定预装包

```bash
sudo ./bin/dp-build build board -n="deepin" -v="beige" -c="main,commercial,community" -a="arm64" -s="deb https://community-packages.deepin.com/beige/ beige main commercial community" -p="deepin-keyring, ca-certificates" -d qemu --image-size 200
```
