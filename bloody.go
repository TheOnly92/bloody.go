package main

import (
	"web"
	"launchpad.net/mgo"
	"os"
	"./mustache"
	"regexp"
	"./session"
)

type User struct {
	Username	string
	Password	string
}

var mSession *mgo.Session
var mongoInit = false
var h *session.MHandler
var config *Config

var layout *mustache.Template

func initMongo() {
	var err os.Error
	mSession, err = mgo.Mongo(config.Get("mongohost"))
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

func pagination(page int, totPages int) map[string]interface{} {
	start := 1
	if page > 3 {
		start = page - 2
	}
	end := start + 4
	if end > totPages {
		end = totPages
	}
	length := 5
	if totPages < length {
		length = totPages
	}
	pages := make([]map[string]int, length)
	cnt := 0
	for i:=start; i <= end; i++ {
		temp := map[string]int{"page": i}
		if i == page {
			temp["current"] = 1
		}
		pages[cnt] = temp
		cnt++
	}
	
	before := true
	beforePage := page - 1
	after := true
	afterPage := page + 1
	lastPage := totPages
	if page == 1 {
		before = false
		beforePage = 0
	}
	if page == totPages {
		after = false
	}
	
	return map[string]interface{} {"Pages": pages, "Before": before, "BeforePage": beforePage, "After": after, "AfterPage": afterPage, "LastPage": lastPage}
}

func main() {
	config = loadConfig()
	initMongo()
	initLayout()
	h = new(session.MHandler)
	h.SetSession(mSession)
	web.Config.StaticDir = config.Get("staticdir")
	web.Get("/", index)
	web.Get("/post/list", listPosts)
	web.Get("/post/([A-Za-z0-9]+)", readPost)
	web.Get("/admin", adminIndexGet)
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