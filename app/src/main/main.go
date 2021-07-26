package main

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"os"
)

var (
	myself dhtNode
	myIP   string
)

func init() {
	var f *os.File
	f, _ = os.Create("log.txt")
	log.SetOutput(f)

	fmt.Println("Please type your IP to quick start :)")
	fmt.Scanln(&myIP)
	fmt.Println("IP is set to ", myIP)
	myself = NewNode(myIP)
	myself.Run()
}
func main() {
	var para1, para2 string
	for {
		fmt.Scanln(&para1, &para2)
		if para1 == "join" {
			ok := myself.Join(para2)
			if ok {
				fmt.Println("Join ", para2, " Successfully!")
			} else {
				fmt.Println("Fail to Join ", para2)
			}
		}
		if para1 == "create" {
			myself.Create()
			fmt.Println("Create new network in ", myIP)
		}
		if para1 == "upload" {
			Lauch(para2, &myself)
		}
		if para1 == "download" {
			download(para2, &myself)
		}
		if para1 == "quit" {
			myself.Quit()
			fmt.Println(myIP, " Node Quit")
		}
		if para1 == "run" {
			myself.Run()
			fmt.Println(myIP, "Run successfully")
		}
	}
}
