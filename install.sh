#!/bin/bash

# CellDock Web 一键安装与系统守护进程注册脚本
# 支持 Linux (树莓派 / Ubuntu / Debian / CentOS) 和 macOS

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo -e "${GREEN}=== 开始安装 CellDock Web 跨平台守护进程 ===${NC}"

# 检查 root 权限
if [ "$EUID" -ne 0 ]; then
    echo -e "${YELLOW}提示: 建议使用 sudo 运行此脚本以注册系统服务${NC}"
fi

INSTALL_DIR="/opt/celldock-web"
mkdir -p "$INSTALL_DIR"

echo -e "${GREEN}1. 正在编译/部署 celldock-web 二进制文件...${NC}"
export PATH=$PATH:/usr/local/go/bin
if command -v go &> /dev/null; then
    go build -o celldock-web main.go
    cp celldock-web "$INSTALL_DIR/celldock-web"
else
    echo -e "${RED}未检测到 Go 编译环境，请先安装 Go 1.20+${NC}"
    exit 1
fi

echo -e "${GREEN}2. 创建守护进程配置文件...${NC}"
if command -v systemctl &> /dev/null; then
    cat <<EOF | sudo tee /etc/systemd/system/celldock-web.service > /dev/null
[Unit]
Description=CellDock Web Cellular Service Daemon
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=/opt/celldock-web
ExecStart=/opt/celldock-web/celldock-web
Restart=always
RestartSec=5
Environment=LISTEN_ADDR=:8080
Environment=DATA_DIR=/opt/celldock-web/data

[Install]
WantedBy=multi-user.target
EOF

    sudo systemctl daemon-reload
    sudo systemctl enable celldock-web
    sudo systemctl restart celldock-web
    echo -e "${GREEN}✓ systemd 守护进程注册成功！${NC}"
fi

echo -e "${GREEN}====================================================${NC}"
echo -e "${GREEN}🎉 CellDock Web 部署成功！${NC}"
echo -e "${GREEN}🌐 请访问 Web 控制台: http://127.0.0.1:8080${NC}"
echo -e "${GREEN}====================================================${NC}"
