package main

import "chord"

func NewNode(ip string) dhtNode {
	ptr := new(chord.Node)
	ptr.Init(ip)
	return ptr
}
