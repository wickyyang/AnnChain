# AnnChain

<img src="https://github.com/dappledger/AnnChain/tree/master/docs/img/AnnChain.png">

AnnChain是众安科技新一代联盟链的核心引擎，拥有集高安全性，高性能，及高可用性等特点于一身，旨在为企业提供紧密集成的区块链系统，非常适用于具有联盟性质的商业机构之间开展业务合作，也适合金融级高频交易、对安全性要求高的场景。目前已应用于技术社区中数十个实际业务场景。

## 特性介绍

安全性：采用PBFT共识算法保障共识结果的一致性，节点准入机制支持CA认证提高安全性。

高性能：支持并行验签，缩短大量交易时的验签时间，提升性能。

易用性：提供部署工具、Golang版和JAVA版SDK、浏览器等配套工具，方便用户使用。

## 版本兼容性说明

0.6.4版本及以下与0.7.1版本及以上数据和协议不兼容，合约兼容。0.6.4版本及以下需升级至0.7.1及以上版本，需要数据迁移。

## 涉及技术

| 技术     | 说明                                 |
| -------- | ------------------------------------ |
| 版本管理 | Git                                  |
| 编程语言 | Golang(版本1.12.0及以上)             |
| 编程工具 | Vscode/JetBrains GoLand/Atom/LiteIDE |
| 操作系统 | CentOS/Ubuntu/MacOS/Windows          |
| 运维工具 | Docker/Docker-compose/Docker-machine |

## 安装

请见[安装说明文档](<https://github.com/dappledger/AnnChain/tree/master/docs/install.md>)

## 快速部署

- [单节点部署说明文档](https://github.com/dappledger/AnnChain/tree/master/docs/singleNode.md)
- docker多节点部署文档

## 执行智能合约

请见[智能合约使用文档](https://github.com/dappledger/AnnChain/tree/master/docs/smartContract.md)

## 使用场景

##### 存证溯源

基于区块链数据不可篡改的特性，可用于数据存证和防伪溯源等场景。

##### 虚拟资产通证化

可用于积分、优惠券、点卡等虚拟资产的发行和流通

##### 实物资产通证化

物理世界的实物资产在区块链中通证化，便于流通交易。比如艺术品、贵金属等高净值实物。

##### 金融资产通证化

金融资产在区块链中通证化，降低交易成本。比如债权、供应链等。

## 区块链知识和行业动态

[安全多方计算的根基 — Yao’s两方协议](http://www.annchain.io/news/juigyBcA-)

[解读Conflux的共识机制](http://www.annchain.io/news/BlkaAVTkL)

[布隆过滤器在区块链中的应用](http://www.annchain.io/news/8tkqctsPf)

[像谈恋爱那样去招区块链工程师](http://www.annchain.io/news/8308r12PE)

[区块链+"UKey"破解上链数据真伪之争，众安携手鼎钻推出钻石标准化交易模式](http://www.annchain.io/news/w-jzv7fBM)

## 联系我们

邮箱： [![](<https://img.shields.io/twitter/url/http/shields.io.svg?logo=Gmail&style=social&label=annchain@zhongan.io>)](mailto:annchain@zhongan.io)

微信群：添加群管理员微信号，拉您进入annchain官方技术交流群。

群管理员微信二维码： [![Scan](https://img.shields.io/badge/style-Scan_QR_Code-green.svg?logo=wechat&longCache=false&style=social&label=Ann)](https://github.com/wickyyang/AnnChain/blob/v0.6.4_stable/doc/annChain.Genesis.png)

如有疑问，欢迎[咨询](https://github.com/dappledger/AnnChain/issues)和[提交BUG](https://github.com/dappledger/AnnChain/issues)。

## License

[![license](<https://img.shields.io/badge/license-Apache--2.0-brightgreen>)](<https://github.com/dappledger/AnnChain/blob/master/LICENSE>)

AnnChain的开源协议为Apache 2.0。详情参见LICENSE。

