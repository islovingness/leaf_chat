package main

import (
	"../common"
	"github.com/islovingness/leaf"
	lconf "github.com/islovingness/leaf/conf"
	"./conf"
	"os"
	"github.com/islovingness/leaf/log"
	"./center"
	"./room"
	"github.com/islovingness/leaf/module"
)

func main() {
	argsLen := len(os.Args)
	if argsLen < 2 {
		log.Fatal("os args of len(%v) less than 2", argsLen)
	}

	confFileName := os.Args[1]
	conf.Init(confFileName)

	lconf.LogLevel = conf.Server.LogLevel
	lconf.LogPath = conf.Server.LogPath
	lconf.LogFlag = conf.LogFlag
	lconf.ConsolePort = conf.Server.ConsolePort
	lconf.ProfilePath = conf.Server.ProfilePath
	lconf.ServerName = conf.Server.ServerName
	lconf.ListenAddr = conf.Server.ListenAddr
	lconf.ConnAddrs = conf.Server.ConnAddrs
	lconf.PendingWriteNum = conf.Server.PendingWriteNum
	lconf.HeartBeatInterval = conf.HeartBeatInterval

	common.Init()

	modules := []module.Module{center.Module}
	modules = append(modules, room.CreateModules()...)
	leaf.Run(modules...)
}
