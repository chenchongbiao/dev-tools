#!/usr/bin/make -f

.PHONY: build install_dep

PREFIX = usr
BINARY_DIR = bin
DP_BUILD_CMD = dp-build
TAG = 1.0.0
GOBUILD = CGO_ENABLED=0 go build -mod vendor -ldflags "-X main.version=$(TAG)" -v $(GO_BUILD_FLAGS)

export GO111MODULE=on

# 构建可执行文件
build:
	${GoPath} ${GOBUILD} -o ${BINARY_DIR}/${DP_BUILD_CMD} ./cmd/${DP_BUILD_CMD}

# 安装一些必要的服务，如果需要在 x86 构建其他架构需要安装 qemu-user-static，并重启 systemd-binfmt 服务
install-dep:
	sudo apt install mmdebstrap qemu-user-static usrmerge systemd-container gdisk
	sudo systemctl restart systemd-binfmt