- 利用`chord协议`实现了命令行交互的分布式文件存储系统。提供了节点加入网络、文件上传、种子生成、据种子下载文件的功能。

- 参考了https://blog.jse.li/posts/torrent/#putting-it-all-together 这篇博客

- 使用了https://github.com/jackpal/bencode-go 开源`bencode`编码解码包

- 参考使用了https://github.com/veggiedefender/torrent-client 的少部分代码，包括：

  - `bencodeInfo`、`bencodeTorrent`、`TorrentFile`结构体定义

  - `func (i *bencodeInfo) splitPieceHashes() ([][20]byte, error)`

    `func (bto *bencodeTorrent) ToTorrentFile() (TorrentFile, error)`

