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
	"os/exec"
	"os"
	//"fmt"
	//"reflect"
)

type Admin struct {}

func (c *Admin) IndexGet(ctx *web.Context) string {
	sessionH := session.Start(ctx, h)
	defer sessionH.Save()
	if sessionH.Data["logged"] != nil {
		ctx.Redirect(302, "/admin/post/list")
		return ""
	}
	ctx.Redirect(302, "/admin/login")
	return ""
}

func (c *Admin) LoginGet(ctx *web.Context) string {
	sessionH := session.Start(ctx, h)
	defer sessionH.Save()
	if sessionH.Data["logged"] != nil {
		ctx.Redirect(302, "/admin/post/list")
		return ""
	}
	output := mustache.RenderFile("templates/admin/login.mustache")
	return render(output, "Login")
}

func (c *Admin) LoginPost(ctx *web.Context) {
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

func (c *Admin) PreferencesGet(ctx *web.Context) string {
	sessionH := session.Start(ctx, h)
	defer sessionH.Save()
	if sessionH.Data["logged"] == nil {
		ctx.Redirect(302, "/admin/login")
		return ""
	}
	
	vars := map[string]string {"DateFormat": blogConfig.Get("dateFormat"), "PostsPerPage": blogConfig.Get("postsPerPage")}
	comment := blogConfig.Get("enableComment")
	if comment != "" {
		vars["EnableComment"] = comment
	}
	if blogConfig.Get("markdown") == "1" {
		vars["markdown"] = "1"
	}
	output := mustache.RenderFile("templates/admin/preferences.mustache", vars)
	return render(output, "Blog Preferernces")
}

func (c *Admin) PreferencesPost(ctx *web.Context) {
	sessionH := session.Start(ctx, h)
	defer sessionH.Save()
	if sessionH.Data["logged"] == nil {
		ctx.Redirect(302, "/admin/login")
		return
	}
	
	blogConfig.Update("dateFormat", ctx.Params["dateFormat"])
	blogConfig.Update("postsPerPage", ctx.Params["postsPerPage"])
	blogConfig.Update("enableComment", ctx.Params["enableComment"])
	markdown := "0"
	if ctx.Params["markdown"] == "on" {
		markdown = "1"
	}
	blogConfig.Update("markdown", markdown)
	ctx.Redirect(302, "/admin/preferences")
}

func (c *Admin) NewPostGet(ctx *web.Context) string {
	sessionH := session.Start(ctx, h)
	defer sessionH.Save()
	if sessionH.Data["logged"] == nil {
		ctx.Redirect(302, "/admin/login")
		return ""
	}
	var output string
	if _, exists := ctx.Params["markdown"]; exists {
		output = mustache.RenderFile("templates/admin/new-post-markdown.mustache")
	} else {
		output = mustache.RenderFile("templates/admin/new-post.mustache")
	}
	return render(output, "New Post")
}

func (c *Admin) NewPostPost(ctx *web.Context) {
	sessionH := session.Start(ctx, h)
	defer sessionH.Save()
	if sessionH.Data["logged"] == nil {
		ctx.Redirect(302, "/admin/login")
		return
	}
	p := PostModelInit()
	var markdown uint
	markdown = 0
	if _, exists := ctx.Params["markdown"]; exists {
		markdown = 1
	}
	p.Create(ctx.Params["title"], ctx.Params["content"], ctx.Params["status"], markdown)
	ctx.Redirect(302, "/admin/post/list")
}

func (c *Admin) ListPost(ctx *web.Context) string {
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
	
	vars := map[string]interface{} {"posts":results}
	if blogConfig.Get("markdown") == "1" {
		vars["markdown"] = "1"
	}
	
	output := mustache.RenderFile("templates/admin/list-post.mustache", vars)
	return render(output, "List Posts")
}

func (c *Admin) EditPostGet(ctx *web.Context, postId string) string {
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
	var output string
	if result.Type == 1 {
		output = mustache.RenderFile("templates/admin/edit-post-markdown.mustache", templateVars)
	} else {
		output = mustache.RenderFile("templates/admin/edit-post.mustache", templateVars)
	}
	
	return render(output, "Edit Post")
}

func (c *Admin) EditPostPost(ctx *web.Context, postId string) {
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

func (c *Admin) DelPost(ctx *web.Context, postId string) {
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

func (c *Admin) BulkActions(ctx *web.Context) {
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

func (c *Admin) NewPageGet(ctx *web.Context) string {
	sessionH := session.Start(ctx, h)
	defer sessionH.Save()
	if sessionH.Data["logged"] == nil {
		ctx.Redirect(302, "/admin/login")
		return ""
	}
	output := mustache.RenderFile("templates/admin/new-page.mustache")
	return render(output, "New Page")
}

func (c *Admin) NewPagePost(ctx *web.Context) {
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

func (c *Admin) ListPagesGet(ctx *web.Context) string {
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

func (c *Admin) EditPageGet(ctx *web.Context, id string) string {
	sessionH := session.Start(ctx, h)
	defer sessionH.Save()
	if sessionH.Data["logged"] == nil {
		ctx.Redirect(302, "/admin/login")
		return ""
	}
	p := PageModelInit()
	result := p.Get(id)
	
	output := mustache.RenderFile("templates/admin/edit-page.mustache", map[string]string {"Title":result.Title, "Content":result.Content, "id":objectIdHex(result.Id.String())})
	return render(output, "Edit Post")
}

func (c *Admin) EditPagePost(ctx *web.Context, id string) {
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

func (c *Admin) DelPage(ctx *web.Context, id string) {
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

func (c *Admin) DelComment(ctx *web.Context, postId string, id string) {
	sessionH := session.Start(ctx, h)
	defer sessionH.Save()
	if sessionH.Data["logged"] == nil {
		ctx.Redirect(302, "/admin/login")
		return
	}
	
	p := PostModelInit()
	p.DeleteComment(postId, id)
	
	ctx.Redirect(302, "/post/"+postId)
}

func (c *Admin) RestartBloody(ctx *web.Context) string {
	sessionH := session.Start(ctx, h)
	defer sessionH.Save()
	if sessionH.Data["logged"] == nil {
		ctx.Redirect(302, "/admin/login")
		return ""
	}
	
	pid := os.Getpid()
	dir, _ := os.Getwd()
	command1 := exec.Command("kill",strconv.Itoa(pid))
	command2 := exec.Command(dir+"/bloody")
	command2.Start()
	command1.Start()
	
	output := mustache.RenderFile("templates/admin/restart.mustache")
	return render(output, "Restarting Bloody")
}