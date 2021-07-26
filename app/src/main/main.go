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
	for {
		var para1, para2, para3, para4 string = "", "", "", ""
		fmt.Scanln(&para1, &para2, &para3, &para4)
		if para1 == "join" {
			ok := myself.Join(para2)
			if ok {
				fmt.Println("Join ", para2, " Successfully!")
			} else {
				fmt.Println("Fail to Join ", para2)
			}
			continue
		}
		if para1 == "create" {
			myself.Create()
			fmt.Println("Create new network in ", myIP)
			continue
		}
		if para1 == "upload" {
			err := Lauch(para2, para3, &myself)
			if err != nil {
				fmt.Println("Fail to upload ", para2)
			}
			continue
		}
		if para1 == "download" {
			if para2 == "-t" {
				err := download(para3, para4, &myself)
				if err != nil {
					fmt.Println("Fail to download ", err)
				}
				continue
			}
			if para2 == "-m" {
				err := downloadByMagnet(para3, para4, &myself)
				if err != nil {
					fmt.Println("Fail to download ", err)
				}
				continue
			}
		}
		if para1 == "quit" {
			myself.Quit()
			fmt.Println(myIP, " Node Quit")
			continue
		}
		if para1 == "run" {
			myself.Run()
			fmt.Println(myIP, "Run successfully")
			continue
		}
		fmt.Println("Unknown instruction :(")
	}
}
