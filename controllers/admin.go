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
	//"fmt"
	//"reflect"
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
	return render(output, "Login")
}

func adminLoginPost(ctx *web.Context) {
	sessionH := session.Start(ctx, h)
	defer sessionH.Save()
	if sessionH.Data["logged"] != nil {
		ctx.Redirect(302, "/admin/post/list")
		return
	}
	if ctx.Params["username"] == config.Get("adminuser") && ctx.Params["password"] == config.Get("adminpasswd") {
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
	return render(output, "New Post")
}

func newPostPost(ctx *web.Context) {
	sessionH := session.Start(ctx, h)
	defer sessionH.Save()
	if sessionH.Data["logged"] == nil {
		ctx.Redirect(302, "/admin/login")
		return
	}
	p := PostModelInit()
	p.Create(ctx.Params["title"], ctx.Params["content"], ctx.Params["status"])
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
	p := PostModelInit()
	results := p.PostListing(page)
	
	output := mustache.RenderFile("templates/list-post.mustache", map[string][]map[string]string {"posts":results})
	return render(output, "List Posts")
}

func editPostGet(ctx *web.Context, postId string) string {
	sessionH := session.Start(ctx, h)
	defer sessionH.Save()
	if sessionH.Data["logged"] == nil {
		ctx.Redirect(302, "/admin/login")
		return ""
	}
	p := PostModelInit()
	result := p.Get(postId)
	
	templateVars := map[string]interface{} {"Title":result.Title, "Content":result.Content, "id":objectIdHex(result.Id.String())}
	if result.Status == 0 {
		templateVars["Draft"] = 1
	}
	if result.Status == 1 {
		templateVars["Publish"] = 1
	}
	output := mustache.RenderFile("templates/edit-post.mustache", templateVars)
	return render(output, "Edit Post")
}

func editPostPost(ctx *web.Context, postId string) {
	sessionH := session.Start(ctx, h)
	defer sessionH.Save()
	if sessionH.Data["logged"] == nil {
		ctx.Redirect(302, "/admin/login")
		return
	}
	p := PostModelInit()
	p.Update(ctx.Params["title"], ctx.Params["content"], ctx.Params["status"], postId)
	
	ctx.Redirect(302, "/admin/post/list")
}

func delPost(ctx *web.Context, postId string) {
	sessionH := session.Start(ctx, h)
	defer sessionH.Save()
	if sessionH.Data["logged"] == nil {
		ctx.Redirect(302, "/admin/login")
		return
	}
	p := PostModelInit()
	p.Delete(postId)
	
	ctx.Redirect(302, "/admin/post/list")
}

func adminBulkActions(ctx *web.Context) {
	sessionH := session.Start(ctx, h)
	defer sessionH.Save()
	if sessionH.Data["logged"] == nil {
		ctx.Redirect(302, "/admin/login")
		return
	}
	p := PostModelInit()
	switch ctx.Params["action"] {
	case "delete":
		p.DeleteBulk(ctx.FullParams["posts[]"])
	case "publish":
		p.PublishBulk(ctx.FullParams["posts[]"])
	}
	
	ctx.Redirect(302, "/admin/post/list")
}

func adminNewPageGet(ctx *web.Context) string {
	sessionH := session.Start(ctx, h)
	defer sessionH.Save()
	if sessionH.Data["logged"] == nil {
		ctx.Redirect(302, "/admin/login")
		return ""
	}
	output := mustache.RenderFile("templates/new-page.mustache")
	return render(output, "New Page")
}

func adminNewPagePost(ctx *web.Context) {
	sessionH := session.Start(ctx, h)
	defer sessionH.Save()
	if sessionH.Data["logged"] == nil {
		ctx.Redirect(302, "/admin/login")
		return
	}
	p := PageModelInit()
	p.Create(ctx.Params["title"], ctx.Params["content"])
	ctx.Redirect(302, "/admin/page/list")
}

func adminListPagesGet(ctx *web.Context) string {
	sessionH := session.Start(ctx, h)
	defer sessionH.Save()
	if sessionH.Data["logged"] == nil {
		ctx.Redirect(302, "/admin/login")
		return ""
	}
	p := PageModelInit()
	results := p.List()
	
	output := mustache.RenderFile("templates/list-pages.mustache", map[string][]map[string]string {"pages":results})
	return render(output, "List Pages")
}

func adminEditPageGet(ctx *web.Context, id string) string {
	sessionH := session.Start(ctx, h)
	defer sessionH.Save()
	if sessionH.Data["logged"] == nil {
		ctx.Redirect(302, "/admin/login")
		return ""
	}
	p := PageModelInit()
	result := p.Get(id)
	
	output := mustache.RenderFile("templates/edit-page.mustache", map[string]string {"Title":result.Title, "Content":result.Content, "id":objectIdHex(result.Id.String())})
	return render(output, "Edit Post")
}

func adminEditPagePost(ctx *web.Context, id string) {
	sessionH := session.Start(ctx, h)
	defer sessionH.Save()
	if sessionH.Data["logged"] == nil {
		ctx.Redirect(302, "/admin/login")
		return
	}
	p := PageModelInit()
	p.Update(ctx.Params["title"], ctx.Params["content"], id)
	
	ctx.Redirect(302, "/admin/page/list")
}

func adminDelPage(ctx *web.Context, id string) {
	sessionH := session.Start(ctx, h)
	defer sessionH.Save()
	if sessionH.Data["logged"] == nil {
		ctx.Redirect(302, "/admin/login")
		return
	}
	p := PageModelInit()
	p.Delete(id)
	
	ctx.Redirect(302, "/admin/page/list")
}