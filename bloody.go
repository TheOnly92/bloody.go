package main

import (
	"web"
	"launchpad.net/mgo"
	"./mustache"
	"regexp"
	"./session"
	"bytes"
	"os"
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
	var err error
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
	path, _ := os.Getwd()
	web.Config.StaticDir = path + "/" + config.Get("staticdir")
	i := &Index{}
	a := &Admin{}
	web.Get("/", web.MethodHandler(i, "Index"))
	web.Get("/post/list", web.MethodHandler(i, "ListPosts"))
	web.Get("/post/([A-Za-z0-9]+)", web.MethodHandler(i, "ReadPost"))
	web.Get("/page/([a-z0-9\\-]+)\\.html", web.MethodHandler(i, "ReadPage"))
	web.Post("/post/([A-Za-z0-9]+)/comment/new", web.MethodHandler(i, "NewComment"))
	web.Get("/admin", web.MethodHandler(a, "IndexGet"))
	web.Get("/admin/preferences", web.MethodHandler(a, "PreferencesGet"))
	web.Post("/admin/preferences", web.MethodHandler(a, "PreferencesPost"))
	web.Get("/admin/post/new", web.MethodHandler(a, "NewPostGet"))
	web.Post("/admin/post/new", web.MethodHandler(a, "NewPostPost"))
	web.Get("/admin/post/list", web.MethodHandler(a, "ListPost"))
	web.Post("/admin/post/list", web.MethodHandler(a, "BulkActions"))
	web.Get("/admin/post/edit/(.*)", web.MethodHandler(a, "EditPostGet"))
	web.Post("/admin/post/edit/(.*)", web.MethodHandler(a, "EditPostPost"))
	web.Get("/admin/post/del/(.*)", web.MethodHandler(a, "DelPost"))
	web.Get("/admin/login", web.MethodHandler(a, "LoginGet"))
	web.Post("/admin/login", web.MethodHandler(a, "LoginPost"))
	web.Get("/admin/page/new", web.MethodHandler(a, "NewPageGet"))
	web.Post("/admin/page/new", web.MethodHandler(a, "NewPagePost"))
	web.Get("/admin/page/list", web.MethodHandler(a, "ListPagesGet"))
	web.Get("/admin/page/edit/(.*)", web.MethodHandler(a, "EditPageGet"))
	web.Post("/admin/page/edit/(.*)", web.MethodHandler(a, "EditPagePost"))
	web.Get("/admin/page/del/(.*)", web.MethodHandler(a, "DelPage"))
	web.Get("/admin/comment/del/(.*)/(.*)", web.MethodHandler(a, "DelComment"))
	web.Get("/admin/bloody/restart", web.MethodHandler(a, "RestartBloody"))
	web.Get("/rss", web.MethodHandler(i, "RSS"))
	web.Run(config.Get("host")+":"+config.Get("port"))
}