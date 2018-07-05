package mongodb

import (
	"github.com/islovingness/leaf/db/mongodb"
	"../../conf"
	"github.com/islovingness/leaf/log"
)

var (
	Context *mongodb.DialContext
)

func init()  {
	var err error
	Context, err = mongodb.Dial(conf.Server.MongodbAddr, conf.Server.MongodbSessionNum)
	if err != nil {
		log.Fatal("mongondb init is error(%v)", err)
	}
}
