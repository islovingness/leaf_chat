package internal

import (
	"github.com/islovingness/leaf/cluster"
	"github.com/islovingness/leaf/log"
	"errors"
	"gopkg.in/mgo.v2/bson"
	"math"
)

var (
	frontInfoMap    = map[string]*FrontInfo{}
	chatInfoMap     = map[string]*ChatInfo{}
	roomInfoMap     = map[string]*RoomInfo{}
	accountFrontMap = map[bson.ObjectId]*FrontInfo{}
)

type FrontInfo struct {
	serverName     string
	clientCount    int
	maxClientCount int
	clientAddr     string
}

type ChatInfo struct {
	serverName  string
	clientCount int
	clusterAddr string
}

type RoomInfo struct {
	serverName string
}

func handleRpc(id interface{}, f interface{}) {
	cluster.SetRoute(id, ChanRPC)
	skeleton.RegisterChanRPC(id, f)
}

func init() {
	cluster.AgentChanRPC = ChanRPC

	skeleton.RegisterChanRPC("NewServerAgent", NewServerAgent)
	skeleton.RegisterChanRPC("CloseServerAgent", CloseServerAgent)

	handleRpc("GetBestFrontInfo", GetBestFrontInfo)
	handleRpc("UpdateFrontInfo", UpdateFrontInfo)
	handleRpc("UpdateChatInfo", UpdateChatInfo)
	handleRpc("GetRoomInfo", GetRoomInfo)
	handleRpc("DestroyRoom", DestroyRoom)
	handleRpc("AccountOffline", AccountOffline)
}

func NewServerAgent(args []interface{}) {
	serverName := args[0].(string)
	agent := args[1].(*cluster.Agent)
	if serverName[:5] == "front" {
		results, err := agent.CallN("GetFrontInfo")
		if err == nil {
			clientCount := results[0].(int)
			maxClientCount := results[1].(int)
			clientAddr := results[2].(string)
			frontInfoMap[serverName] = &FrontInfo{serverName: serverName, clientCount: clientCount,
				maxClientCount:                               maxClientCount, clientAddr: clientAddr}

			if len(chatInfoMap) > 0 {
				serverInfoMap := map[string]string{}
				for chatName, chatInfo := range chatInfoMap {
					serverInfoMap[chatName] = chatInfo.clusterAddr
				}
				agent.Go("AddClusterClient", serverInfoMap)
			}
		} else {
			log.Error("GetFrontInfo is error: %v", err)
		}
	} else if serverName[:4] == "chat" {
		results, err := agent.CallN("GetChatInfo")
		if err == nil {
			clientCount := results[0].(int)
			clusterAddr := results[1].(string)
			chatInfoMap[serverName] = &ChatInfo{serverName: serverName, clientCount: clientCount, clusterAddr: clusterAddr}

			cluster.Broadcast("front", "AddClusterClient", map[string]string{serverName: clusterAddr})
		} else {
			log.Error("GetChatInfo is error: %v", err)
		}
	}
}

func CloseServerAgent(args []interface{}) {
	serverName := args[0].(string)
	if serverName[:5] == "front" {
		_, ok := frontInfoMap[serverName]
		if ok {
			delete(frontInfoMap, serverName)
		}
	} else if serverName[:4] == "chat" {
		_, ok := chatInfoMap[serverName]
		if ok {
			delete(chatInfoMap, serverName)

			cluster.Broadcast("front", "RemoveClusterClient", serverName)
		}
	}
}

func GetBestFrontInfo(args []interface{}) ([]interface{}, error) {
	accountId := args[0].(bson.ObjectId)

	var ok bool
	var frontInfo *FrontInfo
	if frontInfo, ok = accountFrontMap[accountId]; !ok {
		minClientCount := math.MaxInt32
		for _, _frontInfo := range frontInfoMap {
			if _frontInfo.clientCount < minClientCount && _frontInfo.clientCount < _frontInfo.maxClientCount {
				frontInfo = _frontInfo
			}
		}
	}

	if frontInfo == nil {
		return []interface{}{}, errors.New("No front server to alloc")
	} else {
		accountFrontMap[accountId] = frontInfo
		log.Debug("%v account ask front info", accountId)
		return []interface{}{frontInfo.serverName, frontInfo.clientAddr}, nil
	}
}

func UpdateFrontInfo(args []interface{}) {
	serverName := args[0].(string)
	clientCount := args[1].(int)
	frontInfo, ok := frontInfoMap[serverName]
	if ok {
		frontInfo.clientCount = clientCount
		log.Debug("%v server of client count is %v", serverName, clientCount)
	}
}

func UpdateChatInfo(args []interface{}) {
	serverName := args[0].(string)
	clientCount := args[1].(int)
	chatInfo, ok := chatInfoMap[serverName]
	if ok {
		chatInfo.clientCount = clientCount
		log.Debug("%v server of client count is %v", serverName, clientCount)
	}
}

func GetRoomInfo(args []interface{}) (serverName interface{}, err error) {
	roomName := args[0].(string)
	roomInfo, ok := roomInfoMap[roomName]
	if ok {
		serverName = roomInfo.serverName
	} else {
		var chatInfo *ChatInfo
		minClientCount := math.MaxInt32
		for _, _chatInfo := range chatInfoMap {
			if _chatInfo.clientCount < minClientCount {
				chatInfo = _chatInfo
			}
		}

		if chatInfo == nil {
			err = errors.New("No chat server to alloc")
		} else {
			serverName = chatInfo.serverName
			roomInfoMap[roomName] = &RoomInfo{serverName: chatInfo.serverName}
		}
	}
	return
}

func DestroyRoom(args []interface{}) {
	roomNames := args[0].([]string)
	for _, roomName := range roomNames {
		if _, ok := roomInfoMap[roomName]; ok {
			delete(roomInfoMap, roomName)

		}
	}
	log.Debug("%v rooms is destroy", roomNames)
}

func AccountOffline(args []interface{}) {
	accountId := args[0].(bson.ObjectId)
	if _, ok := accountFrontMap[accountId]; ok {
		delete(accountFrontMap, accountId)
		log.Debug("%v account is offline", accountId)
	}
}
