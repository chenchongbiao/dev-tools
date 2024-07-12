# 使用

## 默认安装
```
sudo ./bin/dp-build build board -n="deepin" -v="beige" -c="main,commercial,community" -a="arm64" -s="deb https://community-packages.deepin.com/beige/ beige main commercial community" -d qemu --image-size 200
```

## 指定预装包
```bash
sudo ./bin/dp-build build board -n="deepin" -v="beige" -c="main,commercial,community" -a="arm64" -s="deb https://community-packages.deepin.com/beige/ beige main commercial community" -p="deepin-keyring, ca-certificates" -d qemu --image-size 200
```
