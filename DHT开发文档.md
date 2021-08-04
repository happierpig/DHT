# Distributed Hash Table - PPCA 2021

## 简介

本项目实现了分布式哈希表`chord && kademlia Protocol`，以及基于`chord Protocol`的简易文件分享application。

`分布式哈希表(DHT)`：是一个去中心化的<key,value>哈希表。其中各节点地位相同，无中心服务器，数据空间通过一定的规则被分割存储到不同的节点上，节点之间通过拓扑网络通讯连接以保证健壮性。

DHT减少了系统单点故障的可能性、能将流量负担分摊到各个节点上面、有一定的匿名性

## Chord Protocol

### Node/Key Distribution

- 每一个Node都有`ip:port`通过`SHA1`得到ID值

  `SHA1`一致哈希：有某些特殊性质，不懂

- 每一个要存的东西可以抽象成<key,value> key可以是文件的名字。对`key`进行`SHA1`得到ID值

- 规定<key,value>存储在successor(SHA1(key))当中

  其中`successor(n)`指ID第一个大于等于(clockwise)n的节点

  将`ID值`放在一个圆周上，key就存储在`ID(key)`顺时针后第一个节点

- ID值用`big int`存储(160 bits)

### Efficiency

- `Finger Table`：进行O(logn)的搜索跳转

  - FingerTable[160]：FingerTable[i]存储的是`successor((nodeID+2^i) % 2^(160))`，最远可跳半个圆周
  - 利用`Finger Table`跳表时：从**159->0**找到`targetID`的前继(即jumpID属于(nowID,targetID))。这样即使`FingerTable`有一些错误，也能牺牲一些性能来保障正确性。

- `Successor List`: 保证`successor fail`之后重连接链表的效率

  避免链断开的情况

### Robustness

- Maintain a `successor list` via copying successor’s SUCCESSOR LIST, removing the last entry of it and adding the successor as the first entry of the list.
  - `stabilize()`: check `successor fail`可以将后继改为`successor list`中的下一个`Online Node`，然后修改`successor list`
  - `Periodically notify`新后继 更改前继
- Check whether successor changed and notify the new successor about himself if changed `periodically`.
  - `stablize() and notify()`: this.successor.prev is between this and this.successor
- `Notify` the successor about himself periodically so that the successor can update his predecessor
  - `notify`: if this.prev == nil || notifier is between this and this.prev then change this.prev to notifier

- Fix one finger table entry per fix interval

- Notify the predecessor and successor when a node want to quit
  - `quit()`: modify pre's`successor list` and suc's `pre`
- Check whether predecessor is failed periodically so that **it can be updated to new available predecessor**
  - `check_predecessor()`: if pre fail, `pre`= nil
  - 真正的修改源自于`notify`

- Data backup
  - 每一个节点多保存一份前继的数据
  - 当 前继fail，就将数据塞入自己dataset和后继的backup

### Implementation 

- Robustness functions are periodically run by `goroutine && time.sleep` 

- 各节点通信的模拟用`rpc`实现

  - 节点类应该有一个成员管理网络通信

    `Client()`、`CallFunction()`、`Ping()`、`Launch()`

  - 对节点类进行封装以`Register`

  - `指针`其实是`IP address`，才能构造`client`对象

- goroutine进程中应该有进行完的条件而不能是无限运行，打一个bool值即可

- 用logrus记录程序运行的状态以debug

- `Node`：代表了DHT网络里的一个节点。

  `function`：实现了协议需要的一些工具函数

  `network`：封装了rpc相关的内容，接口有`init(IP string)`服务端初始化、`RemoteCall()`客户端远程调用

  `wrapnode`：实际上就是包了一层结构体，使得需要rpc调用的函数满足指定格式


### Debug

1. channel会阻塞

2. 小心死锁，Lock区块避开递归

3. Register函数参数顺序必须是Input、output，只有input才发送出去，output输出回来

4. 慎用指针，因为并不共享一块内存空间，数据交流依赖于rpc，直接传输对象的字节流。

5. `use of closed connection`：通过自定义server.Accept()规避

   ```go
   func MyAccept(server *rpc.Server, lis net.Listener, ptr *Node) { // used for closing listener
   	for {
   		conn, err := lis.Accept() // block,so goroutine a new thread
   		select {
   		case <-ptr.station.QuitSignal:
   			return
   		default:
   			if err != nil {
   				log.Print("rpc.Serve: accept:", err.Error())
   				return
   			}
   			go server.ServeConn(conn)
   		}
   	}
   }
   ```

6. for 与 select 嵌套使用，break只能跳出select

## Kademlia Protocol

### Documents

> http://www.yeolar.com/note/2010/03/21/kademlia/
>
> https://program-think.blogspot.com/2017/09/Introduction-DHT-Kademlia-Chord.html

### Node/Key Distribution 

- Node/Key 都由 `SHA1()`得到的`[20] byte`也即`160大小的比特串`代表，`type ID [20]byte `。

  Node 组成一个161层的二叉树且都在叶子结点。

  0往左子树走一层，1往右子树走一层

- 距离定义方式：`Xor`。满足对称性、自零性以及三角不等式，性质很好
- <key,value>存储在**距离**ID(key)最近的K个节点当中。

### Efficiency 

- `alpha`：并发参数，迭代函数中同一时间运行最多的线程数目。提高运行速度。

- `RoutingTable[160]`：

  - RoutingTable[i] 代表DHT树上选择第i个路径时与本节点不同的节点，其中挑选出来`K个`节点的信息存储在这里面用作路由。

    `PrefixLen()`：两个ID进行Xor操作再求前缀零的个数便可求得i

  - `KBuckets`：存储那一部分子树当中至多`K`个节点。维护方式：队尾最近访问的，队首最远访问的，有新的元素插入则检查队首是否在线，是则抛弃新元素、将队首放在队尾，否则抛弃对手戏、将新元素放在队尾。这样保证路由表当中节点的有效性
  - 路由方式O(logn)：迭代。从一个点出发利用RoutingTable找最近的K个点，再从这K个点当中找最近的K个点，直到没有更近的节点被找到。

- 数据存储方式与Chord不同。
  - Chord精准的存储在某个节点(以及备份)，所以提供delete功能。
  - Kademlia将数据存储在距离ID(key)最近的K个节点当中，并有`duplicate`、`republic`以及`热cache`使数据扩散到多个节点上。提高了查询速率(**访问越多的文件cache越多**)，避免某些文件源节点的高负荷，同时也失去了delete功能
  - 文件删除通过`expire`过期自动删除

### Robustness 

- Connection in Network：

  - 每一次成功的远程调用都会Update服务端和客户端的RoutingTable

  - `Refresh:` 每一个节点都会主动去刷新自己的RoutingTable，通过调用`FindKClosest()`的方式 by `goroutime + time.sleep`

- Storage：将数据节点分权限划为3级

  - `Publisher`：无`expire`权、无`duplicate`权、有`republic`权

  - `Duplicater`：这些存储都拥有`expire`和`duplicate`的权力，为了保证网络中至少有`K`份文件备份

  - `Common`：二次传播的产物，不拥有传播`duplicate`的能力，提供了热cache的功能。

  `expire`：过期时间应该有合适的算法。**上一级的`store`命令可以覆盖下一级的expire，同级或者逆向则不行。**这样保证了节点可以过期，并且可以续命。

  `republic`：FindKClosest当中投放Duplicater；`duplicater`：FindKClosest当中投放Common

- `goroutine`一直运行进程：`refresh` `duplicate` `expire` `republic`以检查应该执行的操作

### Why UDP?

The main reason is that you rapidly query many nodes that you have never established contact to before and possibly will never see again during a lookup.Kademlia lookups are iterative, i.e. requests won't be forwarded. A forwarding DHT would be more suited to long-standing TCP connections. A large chunk of the traffic consists a short-lived exchange of a request and response between nodes of a network potentially ranging in the millions. The overhead of rapidly establishing thousands of TCP connections would be prohibitive.

### Implementation

- `FindClose`：从RoutingTable当中找到K个相近节点。

  在`prefixLen(Xor(this,target))`对应的以及临近的`buckets`当中找到K个节点并返回

- `FindKClosest`：采用迭代的方式不断逼近，其中有alpha并发
- `FindValue`：与FindKClosest基本相同，增加找到便停止查询和路径上cache对应数据

- `Node` Kad网络中的一个节点以及相关方法

  `WrapNode` 包裹Node节点使得方法符合register规则

  `network` 封装rpc相关

  `routingTable` 封装routingtable相关

  `function` 实现一些工具函数，关于ID和Contact

  `config` 常量参数的定义和结构体定义
  
  `database` 包含时间处理的数据库封装

### Debug

- for select 嵌套使用与break

- 计数器(采用atomic并发安全)在goroutine里使用会延迟

  

## Application

[Document](https://blog.jse.li/posts/torrent/#putting-it-all-together)

### .torrent

根据BitTorrent协议，文件发布者会根据要发布的文件生成提供一个.torrent文件，简称为“种子”。种子文件本质上是文本文件，包含**Tracker信息和文件信息**两部分。两部分信息经由`Bencode`编码而成，从种子文件中获取信息我们需要`decode`。

种子文件本质上是文本文件，包含**Tracker信息和文件信息**两部分。Tracker信息主要是BT下载中需要用到的Tracker服务器的地址和针对Tracker服务器的设置，文件信息是根据对目标文件的计算生成的，计算结果根据[Bencode](https://zh.wikipedia.org/wiki/Bencode)规则进行编码。它的主要原理是需要把提供下载的文件虚拟分成大小相等的块，并把**每个块的索引信息**和**hash验证码**写入种子文件中；所以，种子文件就是被下载文件的“索引”。

DHT: 用户无需连上Tracker就可以下载，节点会在DHT网络中寻找下载同一文件的其他节点并与之通讯，开始下载任务。

### Magnet

磁链的作用就是便于扩散，因为磁链就是一个小小的文本。

磁力链接本身是没什么用的，不管在任何软件的磁链下载中，都必须要先通过磁链得到**种子文件**，再使用种子文件进行常规下载。因为种子文件才有分片信息，文件大小，文件名等必要信息。通过磁力连接可以从DHT网络中获取到对应的种子文件，然后根据种子进行下载。

### Description

本应用提供命令行交互的简易文件分享系统。用户可上传指定路径下的文件到DHT网络当中，应用生成一个种子文件到指定路径并生成一个磁力链接。用户可通过种子文件/磁力链接从DHT网络当中下载文件。节点的加入和退出不影响以上文件。

### Implementation

- `Encode / Decode`：利用 https://github.com/jackpal/bencode-go 工具在种子文件和内存结构体之间相互转换，获取信息/编码信息

- 上传：将文件切分成多个小块，每个小块大小为`pieceLength`，并计算出对应`hash索引值` ，然后利用`pending`、`finished`两个队列并发上传到DHT网络当中，然后生成种子/磁力链接

- 下载：利用磁力链接下载种子文件。`decode`种子文件获取每一个小块的索引值，然后利用`pending`、`finished`两个队列并发下载小块，每一个下载的小块重新计算`hash值`检查完整性，不完整则加入`pending`重新下载。