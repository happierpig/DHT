# Distributed Hash Table - PPCA 2021

## 简介

本项目实现了分布式哈希表`chord && kademlia Protocol`，以及基于`chord Protocol`的简易文件分享application。

## Chord Protocol

- 实现的类

  `Node`：代表了DHT网络里的一个节点，实现了维持Robustness的方法。

  `function`：实现了协议需要的一些工具函数

  `network`：封装了rpc相关的内容，接口有`init(IP string)`服务端初始化、`RemoteCall()`客户端远程调用

  `wrapnode`：实际上就是包了一层结构体，使得需要rpc调用的函数满足指定格式

- 提供接口：

  `Init(port int)` 、`Run()`、 `Create()`、 `Join(address string) bool`、 `Quit()`、 `ForceQuit()`、 `Put(key string,value string) bool`、 `Get(key string) (bool,string)`、 `Delete(key string) bool` 

## Kademlia Protocol

- 实现的类

  `Node` Kad网络中的一个节点以及相关方法

  `WrapNode` 包裹Node节点使得方法符合register规则

  `network` 封装rpc相关，包括服务端监听和远程调用方法

  `routingTable` 封装routingtable相关

  `function` 实现一些工具函数，关于ID和Contact的运算

  `config` 常量参数的定义和结构体定义

  `database` 包含时间处理的数据库封装，并发安全

- 提供接口：

  `Init(port int)`、`Run()`、 `Join(address string) bool`、 `Quit()`、  `Put(key string,value string)`、 `Get(key string)(bool,string)`

## Application 

- 设定IP，加入DHT网络后，实现了上传和下载功能
- 上传：将指定路径下的文件上传到DHT网络当中，并在指定路径下产生种子文件、输出磁力链接
- 下载：通过种子文件/磁力链接将目标文件下载到指定路径下



详细开发文档请见：https://github.com/happierpig/DHT/blob/master/DHT%E5%BC%80%E5%8F%91%E6%96%87%E6%A1%A3.md

