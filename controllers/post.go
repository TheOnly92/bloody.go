package main

import (
	"./mustache"
	"time"
	"launchpad.net/gobson/bson"
	"web"
	"os"
	"./session"
	"crypto/sha1"
	"hash"
	"strconv"
	"encoding/hex"
)

func adminLoginGet(ctx *web.Context) string {
	sessionH := session.Start(ctx, h)
	defer sessionH.Save()
	if sessionH.Data["logged"] != nil {
		ctx.Redirect(301, "/admin/post/list")
		return ""
	}
	output := mustache.RenderFile("templates/admin-login.mustache")
	return layout.Render(map[string]string {"Body": output})
}

func adminLoginPost(ctx *web.Context) {
	sessionH := session.Start(ctx, h)
	defer sessionH.Save()
	if sessionH.Data["logged"] != nil {
		ctx.Redirect(301, "/admin/post/list")
		return
	}
	if ctx.Params["username"] == "admin" && ctx.Params["password"] == "123456" {
		t := time.LocalTime()
		var h hash.Hash = sha1.New()
		h.Write([]byte(strconv.Itoa64(t.Seconds())))
		sessionH.Data["logged"] = hex.EncodeToString(h.Sum())
	}
	ctx.Redirect(301, "/admin/post/list")
}

func newPostGet(ctx *web.Context) string {
	sessionH := session.Start(ctx, h)
	defer sessionH.Save()
	if sessionH.Data["logged"] == nil {
		ctx.Redirect(301, "/admin/login")
		return ""
	}
	output := mustache.RenderFile("templates/new-post.mustache")
	return layout.Render(map[string]string {"Body": output})
}

func newPostPost(ctx *web.Context) {
	sessionH := session.Start(ctx, h)
	defer sessionH.Save()
	if sessionH.Data["logged"] == nil {
		ctx.Redirect(301, "/admin/login")
		return
	}
	c := mSession.DB(dbname).C("posts")
	t := time.LocalTime()
	err := c.Insert(&Post{"", ctx.Params["title"], ctx.Params["content"], t.Seconds()})
	if err != nil {
		panic(err)
	}
	ctx.Redirect(301, "/admin/post/list")
}

func listPost(ctx *web.Context) string {
	sessionH := session.Start(ctx, h)
	defer sessionH.Save()
	if sessionH.Data["logged"] == nil {
		ctx.Redirect(301, "/admin/login")
		return ""
	}
	c := mSession.DB(dbname).C("posts")
	var result *Post
	results := []map[string]string{}
	err := c.Find(nil).Sort(bson.M{"timestamp":-1}).For(&result, func() os.Error {
		t := time.SecondsToLocalTime(result.Timestamp)
		results = append(results, map[string]string {"id":objectIdHex(result.Id.String()), "Title":result.Title, "Date":t.Format("2006 Jan 02 15:04")})
		return nil
	})
	if err != nil {
		panic(err)
	}
	
	output := mustache.RenderFile("templates/list-post.mustache", map[string][]map[string]string {"posts":results})
	return layout.Render(map[string]string {"Body": output})
}

func editPostGet(ctx *web.Context, postId string) string {
	sessionH := session.Start(ctx, h)
	defer sessionH.Save()
	if sessionH.Data["logged"] == nil {
		ctx.Redirect(301, "/admin/login")
		return ""
	}
	c := mSession.DB(dbname).C("posts")
	var result *Post
	err := c.Find(bson.M{"_id": bson.ObjectIdHex(postId)}).One(&result)
	if err != nil {
		panic(err)
	}
	
	output := mustache.RenderFile("templates/edit-post.mustache", map[string]string {"Title":result.Title, "Content":result.Content, "id":objectIdHex(result.Id.String())})
	return layout.Render(map[string]string {"Body": output})
}

func editPostPost(ctx *web.Context, postId string) {
	sessionH := session.Start(ctx, h)
	defer sessionH.Save()
	if sessionH.Data["logged"] == nil {
		ctx.Redirect(301, "/admin/login")
		return
	}
	c := mSession.DB(dbname).C("posts")
	var result *Post
	err := c.Find(bson.M{"_id": bson.ObjectIdHex(postId)}).One(&result)
	if err != nil {
		panic(err)
	}
	err = c.Update(bson.M{"_id":bson.ObjectIdHex(postId)},&Post{bson.ObjectIdHex(postId),ctx.Params["title"],ctx.Params["content"],result.Timestamp})
	if err != nil {
		panic(err)
	}
	ctx.Redirect(301, "/admin/post/list")
}

func delPost(ctx *web.Context, postId string) {
	sessionH := session.Start(ctx, h)
	defer sessionH.Save()
	if sessionH.Data["logged"] == nil {
		ctx.Redirect(301, "/admin/login")
		return
	}
	c := mSession.DB(dbname).C("posts")
	err := c.Remove(bson.M{"_id":bson.ObjectIdHex(postId)})
	if err != nil {
		panic(err)
	}
	ctx.Redirect(301, "/admin/post/list")
}