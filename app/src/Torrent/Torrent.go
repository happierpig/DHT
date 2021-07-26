package Torrent

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"github.com/jackpal/bencode-go"
	"io"
	"io/ioutil"
	"os"
	"time"
)

var (
	PieceSize int
	TimeWait  time.Duration
)

type BencodeInfo struct {
	Pieces      string `bencode:"pieces"`
	PieceLength int    `bencode:"piece length"`
	Length      int    `bencode:"length"`
	Name        string `bencode:"name"`
}

type bencodeTorrent struct {
	Announce string      `bencode:"announce"`
	Info     BencodeInfo `bencode:"info"`
}

type TorrentFile struct {
	Announce    string
	InfoHash    [20]byte
	PieceHashes [][20]byte
	PieceLength int
	Length      int
	Name        string
}

type PieceInfo struct {
	Index int
	Data  []byte
}

func init() {
	PieceSize = 262144
	TimeWait = 3 * time.Second
}

// Open parses a torrent file
func Open(r io.Reader) (*bencodeTorrent, error) {
	bto := bencodeTorrent{}
	err := bencode.Unmarshal(r, &bto)
	if err != nil {
		return nil, err
	}
	return &bto, nil
}

func (i *BencodeInfo) InfoHash() ([20]byte, error) {
	var buf bytes.Buffer
	err := bencode.Marshal(&buf, *i)
	if err != nil {
		return [20]byte{}, err
	}
	h := sha1.Sum(buf.Bytes())
	return h, nil
}

func (i *PieceInfo) Hash() ([20]byte, error) {
	var buf bytes.Buffer
	err := bencode.Marshal(&buf, *i)
	if err != nil {
		return [20]byte{}, err
	}
	h := sha1.Sum(buf.Bytes())
	return h, nil
}

func (i *BencodeInfo) splitPieceHashes() ([][20]byte, error) {
	hashLen := 20 // Length of SHA-1 hash
	buf := []byte(i.Pieces)
	if len(buf)%hashLen != 0 {
		err := fmt.Errorf("Received malformed pieces of length %d", len(buf))
		return nil, err
	}
	numHashes := len(buf) / hashLen
	hashes := make([][20]byte, numHashes)

	for i := 0; i < numHashes; i++ {
		copy(hashes[i][:], buf[i*hashLen:(i+1)*hashLen])
	}
	return hashes, nil
}

func (bto *bencodeTorrent) ToTorrentFile() (TorrentFile, error) {
	infoHash, err := bto.Info.InfoHash()
	if err != nil {
		return TorrentFile{}, err
	}
	pieceHashes, err := bto.Info.splitPieceHashes()
	if err != nil {
		return TorrentFile{}, err
	}
	t := TorrentFile{
		Announce:    bto.Announce,
		InfoHash:    infoHash,
		PieceHashes: pieceHashes,
		PieceLength: bto.Info.PieceLength,
		Length:      bto.Info.Length,
		Name:        bto.Info.Name,
	}
	return t, nil
}

func MakeTorrentFile(fileName string, targetPath string, ch chan string, ch2 chan BencodeInfo, ch3 chan string) error {
	// name(fileName);total length;piece length = 262144 Bytes ;pieces
	fileState, err := os.Stat(fileName)
	if err != nil {
		fmt.Errorf("<MakeTorrentFile> Fail to get file's state")
		return err
	}
	tmp := bencodeTorrent{
		Announce: "DHT looks down on Trackers :)",
		Info: BencodeInfo{
			Pieces:      "",
			PieceLength: PieceSize,
			Length:      int(fileState.Size()),
			Name:        fileState.Name(),
		},
	}
	tmp.Info.Pieces = <-ch
	var f *os.File
	var realFileName string
	if targetPath == "" {
		realFileName = fileState.Name() + ".torrent"
		f, _ = os.Create(realFileName)
	} else {
		realFileName = targetPath + "/" + fileState.Name() + ".torrent"
		f, _ = os.Create(realFileName)
	}
	err = bencode.Marshal(f, tmp)
	if err != nil {
		fmt.Println("Fail to Marshal the info")
		return err
	}
	fmt.Println("Successfully generate .torrent file named ", fileState.Name()+".torrent")
	content, _ := ioutil.ReadFile(realFileName)
	ch2 <- tmp.Info
	ch3 <- string(content)
	return nil
}
