package main

import (
	"Torrent"
	"fmt"
	"io/ioutil"
	"os"
	"time"
)

type lauchPackage struct {
	hashes [20]byte
	index  int
}

type downloadPackage struct {
	data  []byte
	index int
}

func Lauch(fileName string, node *dhtNode) error {
	content, err := ioutil.ReadFile(fileName)
	if err != nil {
		fmt.Errorf("<Lauch> Fail to open the file :(")
		return err
	}
	var fileByteSize = len(content) // the number of bytes in this file
	var blockNum int                // the number of individual piece
	if fileByteSize%Torrent.PieceSize == 0 {
		blockNum = fileByteSize / Torrent.PieceSize
	} else {
		blockNum = fileByteSize/Torrent.PieceSize + 1
	}

	var pieces []byte = make([]byte, 20*blockNum) // slice make before use
	ch1 := make(chan int, blockNum+20)            // store pieces to be uploaded
	ch2 := make(chan lauchPackage, blockNum+20)
	for i := 1; i <= blockNum; i++ {
		ch1 <- i
	}
	var flag1, flag2 bool = true, true
	for flag1 {
		select { // break has no use
		case index := <-ch1:
			l := (index - 1) * Torrent.PieceSize
			r := index * Torrent.PieceSize
			if r > fileByteSize {
				r = fileByteSize
			}
			go uploadToNetwork(blockNum, node, index, content[l:r], ch1, ch2)
		case <-time.After(Torrent.TimeWait):
			fmt.Println("Upload Finish :)")
			flag1 = false
		}
		time.Sleep(500 * time.Millisecond)
	}
	ch3 := make(chan string)
	go Torrent.MakeTorrentFile(fileName, ch3)
	for flag2 {
		select {
		case pack := <-ch2:
			index := pack.index
			copy(pieces[(index-1)*20:index*20], pack.hashes[:])
		default:
			fmt.Println("Uploading Finished :)")
			flag2 = false
		}
	}
	ch3 <- string(pieces)
	return nil
}

func uploadToNetwork(totalSize int, node *dhtNode, index int, data []byte, ch1 chan int, ch2 chan lauchPackage) {
	info := Torrent.PieceInfo{index, data}
	hashKey, err := info.Hash()
	if err != nil {
		fmt.Println("Fail to hash when upload due to ", err, " Try again :(")
		ch1 <- index // repeat to wait queue
		return
	}
	var sx16 string = fmt.Sprintf("%x", hashKey) // [20]byte->string
	ok := (*node).Put(sx16, string(data))
	if !ok {
		fmt.Println("Fail to upload Try again :(")
		ch1 <- index
		return
	}
	ch2 <- lauchPackage{hashKey, index}
	fmt.Println("Uploading ", float64(totalSize-len(ch1))/float64(totalSize)*100, "% ....")
	return
}

func download(torrentName string, node *dhtNode) error {
	torrentFile, err := os.Open(torrentName)
	if err != nil {
		fmt.Println("Fail to open the .torrent file when try downloading")
		return err
	}
	bT, err := Torrent.Open(torrentFile)
	if err != nil {
		fmt.Println("Fail to Unmarshal the .torrent file")
		return err
	}
	allInfo, err := bT.ToTorrentFile()
	if err != nil {
		fmt.Println("Fail to transform bt")
		return err
	}
	fmt.Println("Start downloading ", allInfo.Name, "  :)")
	var content []byte = make([]byte, allInfo.Length)
	var blockSize int = len(allInfo.PieceHashes)
	ch1 := make(chan int, blockSize+5)
	ch2 := make(chan downloadPackage, blockSize+5)
	for i := 1; i <= blockSize; i++ {
		ch1 <- i
	}
	flag1, flag2 := true, true
	for flag1 {
		select {
		case index := <-ch1:
			go downloadFromNetwork(blockSize, node, allInfo.PieceHashes[index-1], index, ch1, ch2)
		case <-time.After(Torrent.TimeWait):
			fmt.Println("Download and Verify Finished ")
			flag1 = false
		}
		time.Sleep(500 * time.Millisecond)
	}
	for flag2 {
		select {
		case pack := <-ch2:
			l := allInfo.PieceLength * (pack.index - 1)
			r := allInfo.PieceLength * (pack.index)
			if r > allInfo.Length {
				r = allInfo.Length
			}
			copy(content[l:r], pack.data[:])
		default:
			err2 := ioutil.WriteFile(allInfo.Name, content, 0644)
			if err2 != nil {
				fmt.Println("Fail to write in file due to ", err2)
				return err2
			}
			flag2 = false
		}
	}
	return nil
}

func downloadFromNetwork(totalSize int, node *dhtNode, hashSet [20]byte, index int, ch1 chan int, ch2 chan downloadPackage) {
	var sx16 string = fmt.Sprintf("%x", hashSet)
	ok, rawData := (*node).Get(sx16)
	if !ok {
		fmt.Println("Fail to Download Try again :(")
		ch1 <- index
		return
	}
	info := Torrent.PieceInfo{index, []byte(rawData)}
	verifyHash, err := info.Hash()
	if err != nil {
		fmt.Println("Fail to hash when verify due to ", err, " Try again :(")
		ch1 <- index // repeat to wait queue
		return
	}
	if verifyHash != hashSet {
		fmt.Println("Fail to Verify Try again :(")
		ch1 <- index
		return
	}
	ch2 <- downloadPackage{[]byte(rawData), index}
	fmt.Println("Downloading ", float64(totalSize-len(ch1))/float64(totalSize)*100, "% ....")
	return
}
