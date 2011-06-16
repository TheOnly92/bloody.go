package main

import (
	"web"
	"launchpad.net/mgo"
	"os"
	"./mustache"
	"regexp"
	"./session"
	"bytes"
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

var blogConfig *PreferenceModel

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

func render(output string, title string) string {
	vars := make(map[string]interface{})
	vars["Body"] = output
	if title != "" {
		vars["Title"] = map[string]string {"Name": title}
	}
	p := PageModelInit()
	vars["SidePages"] = p.Sidebar()
	return layout.Render(vars)
}

func objectIdHex(objectId string) string {
	var rx_objecthex = regexp.MustCompile("ObjectIdHex\\(\"([A-Za-z0-9]+)\"\\)")
	match := rx_objecthex.FindStringSubmatch(objectId)
	return match[1]
}

func toAscii(str string) string {
	var rx_ascii = regexp.MustCompile("[^a-zA-Z0-9/_|+ \\-]")
	var rx_chars = regexp.MustCompile("[/_|+ \\-]+")
	rt := rx_ascii.ReplaceAllString(str, "")
	rt = string(bytes.ToLower([]byte(rt)))
	rt = rx_chars.ReplaceAllString(rt, "-")
	return rt
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
	blogConfig = PreferenceInit()
	web.Config.StaticDir = config.Get("staticdir")
	web.Get("/", index)
	web.Get("/post/list", listPosts)
	web.Get("/post/([A-Za-z0-9]+)", readPost)
	web.Get("/page/([a-z0-9\\-]+)\\.html", readPage)
	web.Get("/admin", adminIndexGet)
	web.Get("/admin/preferences", adminPreferencesGet)
	web.Post("/admin/preferences", adminPreferencesPost)
	web.Get("/admin/post/new", newPostGet)
	web.Post("/admin/post/new", newPostPost)
	web.Get("/admin/post/list", listPost)
	web.Post("/admin/post/list", adminBulkActions)
	web.Get("/admin/post/edit/(.*)", editPostGet)
	web.Post("/admin/post/edit/(.*)", editPostPost)
	web.Get("/admin/post/del/(.*)", delPost)
	web.Get("/admin/login", adminLoginGet)
	web.Post("/admin/login", adminLoginPost)
	web.Get("/admin/page/new", adminNewPageGet)
	web.Post("/admin/page/new", adminNewPagePost)
	web.Get("/admin/page/list", adminListPagesGet)
	web.Get("/admin/page/edit/(.*)", adminEditPageGet)
	web.Post("/admin/page/edit/(.*)", adminEditPagePost)
	web.Get("/admin/page/del/(.*)", adminDelPage)
	web.Run(config.Get("host")+":"+config.Get("port"))
}