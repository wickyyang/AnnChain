# AnnChain智能合约使用文档

本文档为AnnChain智能合约使用文档，不涉及AnnChain的安装、节点的部署和启动。本说明使用gtool进行智能合约的部署和执行，还可使用Go/Java SDK来执行合约相关操作。

## 创建账户

生成私钥和地址

```
cd AnnChain
./build/gtool account create

privkey: C579D84396CC7D425AFD5ED700140ECA3A0EF9D7E6FB007C4C09CBDE0359D6AF
address: 771403C283A3F46CDA462F7AEFF5DFD28B00F106
```

## 创建合约

执行智能合约相关操作之前，需先启动节点。以下操作默认节点已启动。

##### 执行命令

```
gtool --backend <已启动的validator节点IP地址:RPC端口> evm create --abif <合约abi文件路径> --callf <合约读取调用的json文件路径> --nonce <读取合约的账户nonce>
```

除以上参数，还需在提示"Privkey for user"时输入已经运行的链中的validator节点的私钥privkey。

##### 返回结果

```
contract address 合约地址
tx result 交易hash
```

##### 示例

```
cd AnnChain
./build/gtool --backend "tcp://127.0.0.1:46657" evm create --abif ./scripts/examples/evm/sample.abi --callf ./scripts/examples/evm/sample.json --nonce 0
Privkey for user : 
C579D84396CC7D425AFD5ED700140ECA3A0EF9D7E6FB007C4C09CBDE0359D6AF
contract address: 0xAe119075bd77dE2d8e32629bdb439D967A1EcFe6									
tx result: 0x3121cda109485a5478cb5ff227f8699dd6fa76a69869cc42a12b1b32b9c4b885
```

sample.abi 合约ABI

```abi
[
        {
                "constant": false,
                "inputs": [
                        {
                                "name": "Id",
                                "type": "uint256"
                        },
                        {
                                "name": "Amount",
                                "type": "uint256"
                        }
                ],
                "name": "createCheckInfos",
                "outputs": [],
                "payable": false,
                "stateMutability": "nonpayable",
                "type": "function"
        },
        {
                "constant": true,
                "inputs": [
                        {
                                "name": "Id",
                                "type": "uint256"
                        }
                ],
                "name": "getPremiumInfos",
                "outputs": [
                        {
                                "name": "",
                                "type": "uint256"
                        }
                ],
                "payable": false,
                "stateMutability": "view",
                "type": "function"
        },
        {
                "anonymous": false,
                "inputs": [
                        {
                                "indexed": false,
                                "name": "Id",
                                "type": "uint256"
                        },
                        {
                                "indexed": false,
                                "name": "Amount",
                                "type": "uint256"
                        }
                ],
                "name": "InputLog",
                "type": "event"
        }
]
```

sample.json  合约读取调用的json文件

```json
{
  "bytecode" : "6060604052341561000f57600080fd5b6101818061001e6000396000f30060606040526004361061004c576000357c0100000000000000000000000000000000000000000000000000000000900463ffffffff168063a6226f2114610051578063b051c1e01461007d575b600080fd5b341561005c57600080fd5b61007b60048080359060200190919080359060200190919050506100b4565b005b341561008857600080fd5b61009e6004808035906020019091905050610136565b6040518082815260200191505060405180910390f35b60007fb45ab3e8c50935ce2fa51d37817fd16e7358a3087fd93a9ac7fbddb22a926c358383604051808381526020018281526020019250505060405180910390a1828160000181905550818160010181905550806000808581526020019081526020016000206000820154816000015560018201548160010155905050505050565b60008060008381526020019081526020016000206001015490509190505600a165627a7a723058207eaf119132cfc4008c97339b874c4c16d20d27a72875e55a6a22a29fee30876d0029",																										
  "params" :[]																					 
}
```

| 参数     | 含义         |
| -------- | ------------ |
| bytecode | 合约bytecode |
| params   | 调用参数     |

## 执行合约

##### 执行命令

```
gtool --backend <已启动的validator节点IP地址:RPC端口> evm execute --abif <合约abi文件路径> --callf <合约读取调用的json文件路径> --nonce <读取合约的账户nonce>
```

除以上参数，还需在提示"Privkey for user"时输入已经运行的链中的validator节点的私钥privkey。

##### 返回结果

```
tx result 交易hash
```

##### 示例

```
./build/gtool --backend "tcp://127.0.0.1:46657" evm call --abif ./scripts/examples/evm/sample.abi --callf ./scripts/examples/evm/sample_execute.json --nonce 1
Privkey for user : 
C579D84396CC7D425AFD5ED700140ECA3A0EF9D7E6FB007C4C09CBDE0359D6AF
tx result: 0x2b41d9c05a7be5b85586c53b5a2d3cacc1ded323a18f1c62c51bc2aea0953b55
```

sample_execute.json  合约读取调用的json文件

```json
{
  "contract" : "0xAe119075bd77dE2d8e32629bdb439D967A1EcFe6",		
  "function" : "createCheckInfos",															
  "params":[																									
    1, 100
  ]
}
```

| 参数     | 含义     |
| -------- | -------- |
| contract | 合约地址 |
| function | 调用方法 |
| params   | 调用参数 |

## 读取合约

##### 执行命令

```
gtool --backend <已启动的validator节点IP地址:RPC端口> evm read --abif <合约abi文件路径> --callf <合约读取调用的json文件路径>
```

除以上参数，还需在提示"Privkey for user"时输入已经运行的链中的validator节点的私钥privkey。

##### 返回结果

```
parse result 结果类型 结果数据
```

##### 示例

```
./build/gtool --backend "tcp://127.0.0.1:46657" evm read --abif ./scripts/examples/evm/sample.abi --callf ./scripts/examples/evm/sample_read.json
Privkey for user : 
C579D84396CC7D425AFD5ED700140ECA3A0EF9D7E6FB007C4C09CBDE0359D6AF
parse result: *big.Int 100
```

sample_read.json  合约读取调用的json文件

```json
{
  "contract" : "0xAe119075bd77dE2d8e32629bdb439D967A1EcFe6",	
  "function" : "getPremiumInfos",															
  "params":[																									
    1
  ]
}
```

| 参数     | 含义     |
| -------- | -------- |
| contract | 合约地址 |
| function | 调用方法 |
| params   | 调用参数 |

## 查询Nonce

##### 执行命令

```
gtool --backend <已启动的validator节点IP地址:RPC端口> query nonce --address <要查询的账户地址>
```

##### 返回结果

```
query result nonce数
```

##### 示例

```
./build/gtool --backend "tcp://127.0.0.1:46657" query nonce --address 771403c283a3f46cda462f7aeff5dfd28b00f106
query result: 2
```

## 查询收据Receipt

##### 执行命令

```
gtool --backend <已启动的validator节点IP地址:RPC端口> query receipt --hash <要查询的交易hash>
```

##### 返回结果

```
query result receipt结构体
```

##### 示例

```
./build/gtool --backend "tcp://127.0.0.1:46657" query receipt --hash 0x2b41d9c05a7be5b85586c53b5a2d3cacc1ded323a18f1c62c51bc2aea0953b55
query result: {"root":null,"status":1,"cumulativeGasUsed":21656,"logsBloom":"0x00000000000000000000000000800000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000001000000000000000000000000000000000000000000008000000000000000000000000000000000000000000000000000000000000200000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000002000000000000000000000020000000000000000000000000000000000000000000000000000000000000","logs":[{"address":"0xae119075bd77de2d8e32629bdb439d967a1ecfe6","topics":["0xb45ab3e8c50935ce2fa51d37817fd16e7358a3087fd93a9ac7fbddb22a926c35"],"data":"0x00000000000000000000000000000000000000000000000000000000000000010000000000000000000000000000000000000000000000000000000000000064","blockNumber":"0x64e","transactionHash":"0x2b41d9c05a7be5b85586c53b5a2d3cacc1ded323a18f1c62c51bc2aea0953b55","transactionIndex":"0x0","blockHash":"0x000000000000000000000000ec83a146ca731fdffe4bef69ad260d7389732e87","logIndex":"0x0","removed":false}],"transactionHash":"0x2b41d9c05a7be5b85586c53b5a2d3cacc1ded323a18f1c62c51bc2aea0953b55","contractAddress":"0x0000000000000000000000000000000000000000","gasUsed":21656}
```