package main

import (
	"./mustache"
	"time"
	"web"
	"./session"
	"crypto/sha1"
	"hash"
	"strconv"
	"encoding/hex"
)

func adminIndexGet(ctx *web.Context) string {
	sessionH := session.Start(ctx, h)
	defer sessionH.Save()
	if sessionH.Data["logged"] != nil {
		ctx.Redirect(302, "/admin/post/list")
		return ""
	}
	ctx.Redirect(302, "/admin/login")
	return ""
}

func adminLoginGet(ctx *web.Context) string {
	sessionH := session.Start(ctx, h)
	defer sessionH.Save()
	if sessionH.Data["logged"] != nil {
		ctx.Redirect(302, "/admin/post/list")
		return ""
	}
	output := mustache.RenderFile("templates/admin-login.mustache")
	return layout.Render(map[string]interface{} {"Body": output, "Title": map[string]string {"Name": "Login"}})
}

func adminLoginPost(ctx *web.Context) {
	sessionH := session.Start(ctx, h)
	defer sessionH.Save()
	if sessionH.Data["logged"] != nil {
		ctx.Redirect(302, "/admin/post/list")
		return
	}
	if ctx.Params["username"] == "admin" && ctx.Params["password"] == "123456" {
		t := time.LocalTime()
		var h hash.Hash = sha1.New()
		h.Write([]byte(strconv.Itoa64(t.Seconds())))
		sessionH.Data["logged"] = hex.EncodeToString(h.Sum())
	}
	ctx.Redirect(302, "/admin/post/list")
}

func newPostGet(ctx *web.Context) string {
	sessionH := session.Start(ctx, h)
	defer sessionH.Save()
	if sessionH.Data["logged"] == nil {
		ctx.Redirect(302, "/admin/login")
		return ""
	}
	output := mustache.RenderFile("templates/new-post.mustache")
	return layout.Render(map[string]interface{} {"Body": output, "Title": map[string]string {"Name": "New Post"}})
}

func newPostPost(ctx *web.Context) {
	sessionH := session.Start(ctx, h)
	defer sessionH.Save()
	if sessionH.Data["logged"] == nil {
		ctx.Redirect(302, "/admin/login")
		return
	}
	p := PostModelInit(mSession.DB(config.Get("mongodb")).C("posts"))
	p.Create(ctx.Params["title"], ctx.Params["content"])
	ctx.Redirect(302, "/admin/post/list")
}

func listPost(ctx *web.Context) string {
	sessionH := session.Start(ctx, h)
	defer sessionH.Save()
	if sessionH.Data["logged"] == nil {
		ctx.Redirect(302, "/admin/login")
		return ""
	}
	page := 0
	if temp, exists := ctx.Params["page"]; exists {
		page, _ = strconv.Atoi(temp)
	}
	p := PostModelInit(mSession.DB(config.Get("mongodb")).C("posts"))
	results := p.PostListing(page)
	
	output := mustache.RenderFile("templates/list-post.mustache", map[string][]map[string]string {"posts":results})
	return layout.Render(map[string]interface{} {"Body": output, "Title": map[string]string {"Name": "List Posts"}})
}

func editPostGet(ctx *web.Context, postId string) string {
	sessionH := session.Start(ctx, h)
	defer sessionH.Save()
	if sessionH.Data["logged"] == nil {
		ctx.Redirect(302, "/admin/login")
		return ""
	}
	p := PostModelInit(mSession.DB(config.Get("mongodb")).C("posts"))
	result := p.Get(postId)
	
	output := mustache.RenderFile("templates/edit-post.mustache", map[string]string {"Title":result.Title, "Content":result.Content, "id":objectIdHex(result.Id.String())})
	return layout.Render(map[string]interface{} {"Body": output, "Title": map[string]string {"Name": "Edit Post"}})
}

func editPostPost(ctx *web.Context, postId string) {
	sessionH := session.Start(ctx, h)
	defer sessionH.Save()
	if sessionH.Data["logged"] == nil {
		ctx.Redirect(302, "/admin/login")
		return
	}
	p := PostModelInit(mSession.DB(config.Get("mongodb")).C("posts"))
	p.Update(ctx.Params["title"], ctx.Params["content"], postId)
	
	ctx.Redirect(302, "/admin/post/list")
}

func delPost(ctx *web.Context, postId string) {
	sessionH := session.Start(ctx, h)
	defer sessionH.Save()
	if sessionH.Data["logged"] == nil {
		ctx.Redirect(302, "/admin/login")
		return
	}
	p := PostModelInit(mSession.DB(config.Get("mongodb")).C("posts"))
	p.Delete(postId)
	
	ctx.Redirect(302, "/admin/post/list")
}