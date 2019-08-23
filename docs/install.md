# AnnChain 安装说明文档

本说明文档主要介绍AnnChain的安装，包含环境部署、机器配置介绍、编译源码和启动单节点，不涉及多节点的部署以及启动后的链相关操作。

## 配置环境与工具

本说明以在CentOS7操作系统上为例。

### 机器配置

| 参数 | 推荐配置 | 最低配置 |
| ---- | -------- | -------- |
| CPU  | 2.4GHz   | 1.5GHz   |
| 内存 | 4GB      | 1GB      |
| 核数 | 4核      | 2核      |
| 带宽 | 5Mb      | 1Mb      |

### 软件工具

- 版本管理工具Git

  `yum install git`

- Golang

  - [ ] [安装](https://golang.org/doc/install)

    `yum install go` 

    或

    ```
    https://dl.google.com/go/go1.12.9.linux-amd64.tar.gz
    tar -C /usr/local -xzf go1.12.9.linux-amd64.tar.gz
    ```

  - [ ] 环境配置

    ```
    mkdir .gopkgs
    vi /etc/profile
    export GOROOT=/usr/local/go //设置为go安装的路径
    GOPATH=$HOME/.gopkgs
    export PATH=$PATH:$GOROOT/bin:$GOPATH/bin
    source /etc/profile
    ```

  - [ ] 查看Go版本

    ```
    go version
    go version go1.12.9 linux/amd64
    ```

- Docker容器和Docker-compose工具

## Clone项目与编译

### Clone项目代码

```
git clone https://github.com/dappledger/AnnChain.git
```

### 下载依赖

```
cd AnnChain
./get_pkgs.sh
```

### 编译项目

`make`

编译成功，在./build目录中看到生成的二进制命令文件。

## 链节点部署与启动

本说明文档仅简单介绍单节点的部署与启动。详细描述请见[单节点部署说明文档](https://github.com/dappledger/AnnChain/tree/master/docs/singleNode.md)与docker多节点部署文档。

### 初始化节点

```
./build/genesis init

Log dir is:  ./
Initialized chain_id: genesis-SyaIbH genesis_file: /root/.genesis/genesis.json priv_validator: /root/.genesis/priv_validator.json
Check the files generated, make sure everything is OK.
```

默认节点配置文件路径`/root/.genesis`

### 启动节点

```
./build/genesis run

node (genesis-SyaIbH) is running on 127.0.0.1:46656 ......
```

