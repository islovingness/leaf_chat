package main

import (
	"../common"
	"../common/msg"
	"./client"
	"github.com/islovingness/leaf"
)

func main()  {
	common.Init()
	client.Init(msg.Processor)

	leaf.Run(
		client.Module,
	)
}
