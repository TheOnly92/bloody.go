package main

import (
	"web"
	"launchpad.net/mgo"
	"os"
	"./mustache"
	"launchpad.net/gobson/bson"
	"regexp"
	"./session"
)

var (
	host = "localhost"
	dbname = "bloody"
)

type User struct {
	Username	string
	Password	string
}

type Post struct {
	Id			bson.ObjectId "_id/c"
	Title		string
	Content		string
	Timestamp	int64
}

var mSession *mgo.Session
var mongoInit = false
var h *session.MHandler

var layout *mustache.Template

func initMongo() {
	var err os.Error
	mSession, err = mgo.Mongo(host)
	if err != nil {
		panic(err)
	}
}

func initLayout() {
	layout, _ = mustache.ParseFile("templates/layout.mustache")
}

func objectIdHex(objectId string) string {
	var rx_objecthex = regexp.MustCompile("ObjectIdHex\\(\"([A-Za-z0-9]+)\"\\)")
	match := rx_objecthex.FindStringSubmatch(objectId)
	return match[1]
}

func main() {
	initMongo()
	initLayout()
	h = new(session.MHandler)
	h.SetSession(mSession)
	web.Config = &web.ServerConfig{"./static","0.0.0.0",9999,"98uarpouaskdjiu4231",true}
	web.Get("/", index)
	web.Get("/admin/post/new", newPostGet)
	web.Post("/admin/post/new", newPostPost)
	web.Get("/admin/post/list", listPost)
	web.Get("/admin/post/edit/(.*)", editPostGet)
	web.Post("/admin/post/edit/(.*)", editPostPost)
	web.Get("/admin/post/del/(.*)", delPost)
	web.Get("/admin/login", adminLoginGet)
	web.Post("/admin/login", adminLoginPost)
	web.Run("0.0.0.0:9999")
}