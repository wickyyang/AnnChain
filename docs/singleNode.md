# AnnChain单节点部署说明文档

本说明文档仅详细介绍单节点的部署与启动。多节点部署请见docker多节点部署文档。

## 初始化节点

```
cd AnnChain
./build/genesis init

Log dir is:  ./
Initialized chain_id: genesis-SyaIbH genesis_file: /root/.genesis/genesis.json priv_validator: /root/.genesis/priv_validator.json
Check the files generated, make sure everything is OK.
```

如需指定配置文件路径：

```
--runtime //指定配置文件路径，默认路径为/root/.genesis
```

初始化目录结构

```
.
├── config.toml
├── data
│   ├── archive
│   ├── archive.db
│   ├── blockstore.db
│   ├── chaindata
│   ├── cs.wal
│   ├── evm.db
│   ├── query_cache
│   ├── refuse_list.db
│   ├── state.db
│   └── votechannel.db
├── genesis.json
├── priv_validator.json
└── priv_validator.json.bak
```

## 配置文件

### config.toml

节点启动的配置信息，负责指定启动节点时需要侦听的端口、使用的数据库类型、日志目录、是否启用签名等信息。各参数具体含义如下：

```
app_name = "evm"                  
auth_by_ca = true									
block_size = 5000									
db_backend = "leveldb"						
environment = "production"				
fast_sync = true									
log_path = ""											
moniker = "anonymous"							
non_validator_auth_by_ca = false	
non_validator_node_auth = false
p2p_laddr = "tcp://0.0.0.0:46656"	
rpc_laddr = "tcp://0.0.0.0:46657"
seeds = ""												
signbyca = ""											
skip_upnp = true									
threshold_blocks = 0							
tracerouter_msg_ttl = 5					
```

| 参数                     | 含义                                                         |
| ------------------------ | ------------------------------------------------------------ |
| app_name                 | 指定app                                                      |
| auth_by_ca               | 加入链网络时是否使用CA认证                                   |
| block_size               | 支持最大交易数                                               |
| db_backend               | 底层数据库                                                   |
| environment              | 日志级别，支持development和production                        |
| fast_sync                | 是否启动快速同步                                             |
| log_path                 | 日志路径                                                     |
| moniker                  | 暂不支持修改                                                 |
| non_validator_auth_by_ca | auth_by_ca=true 时有效，表示非验证节点加入链网络时是否使用CA认证。 |
| non_validator_node_auth  | 暂不支持修改                                                 |
| p2p_laddr                | 监听端口                                                     |
| rpc_laddr                | 本地RPC命令监听端口                                          |
| seeds                    | 在节点启动时需要连接的seeds节点，以获取当前链的状态。        |
| signbyca                 | auth_by_ca=true 时有效，CA节点给当前节点公钥的签名。         |
| skip_upnp                | 是否跳过skip_upnp地址映射机制                                |
| threshold_blocks         | 块数据归档门槛，当本地存储的块数据个数达到这个阈值，将触发一次数据归档操作。 |
| tracerouter_msg_ttl      | 暂不支持修改                                                 |

### genesis.json

指定创世区块的配置信息，与以太坊中的gensis.json作用类似。各参数具体含义如下：

```json
{
    "app_hash": "",																	
    "chain_id": "genesis-SyaIbH",										
    "genesis_time": "0001-01-01T00:00:00.000Z",			
    "plugins": "adminOp,querycache",								
    "validators": [
        {
            "amount": 100,													
            "is_ca": true,												
            "pub_key": [														
                1,
                "4A9B150C00317985291591D589523FF7CAD4EFEB902686E2CE67932C3DE389EF"
            ]
        }
    ]
}
```

| 参数         | 含义                                 |
| ------------ | ------------------------------------ |
| app_hash     | 自定义起始的state状态                |
| chain_id     | 链ID                                 |
| genesis_time | 创世                                 |
| plugins      | 支持的插件                           |
| validators   | 节点信息                             |
| amount       | 权重                                 |
| is_ca        | 是否是CA节点，auth_by_ca=true 时有效 |
| pub_key      | 公钥                                 |

### priv_validator.json

指定validator节点的配置信息，在AnnChain中，节点有两种类型：non-validator 和 validator，其中只有 validator 类型的节点会参与共识。各参数具体含义如下：

```json
{
    "address": "2C6D2EE6F2BEB31D93D67C21BB7E331D548326FD",	
    "last_height": 0,																				
    "last_round": 0,																			
    "last_signature": null,																	
    "last_signbytes": "",																		
    "last_step": 0,																					
    "priv_key": [																					
        1,
 "CE3D292AF3519B06E47CFB8545B67DAAED5850766BC81EC2316700F1AC03AAEC4A9B150C00317985291591D589523FF7CAD4EFEB902686E2CE67932C3DE389EF"
    ],
    "pub_key": [																						
        1,
        "4A9B150C00317985291591D589523FF7CAD4EFEB902686E2CE67932C3DE389EF"
    ]
}
```

| 参数           | 含义               |
| -------------- | ------------------ |
| address        | 地址               |
| last_height    | 共识状态，无需修改 |
| last_round     | 共识状态，无需修改 |
| last_signature | 共识状态，无需修改 |
| last_signbytes | 共识状态，无需修改 |
| last_step      | 共识状态，无需修改 |
| priv_key       | 私钥               |
| pub_key        | 公钥               |

## 启动节点

```
cd AnnChain
./build/genesis run

node (genesis-SyaIbH) is running on 127.0.0.1:46656 ......
```

其他选项：

```
--runtime string //指定配置文件路径，默认路径为/root/.genesis
--log_path string //指定链日志路径，默认为./output.log
```

## 重置节点

清除节点数据，高度为0。

```
cd AnnChain
./build/genesis reset

Reset PrivValidator file /alidata1/admin/annchainNode/priv_validator.json
```