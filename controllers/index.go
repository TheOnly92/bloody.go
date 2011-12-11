package main

import (
	"mustache"
	"time"
	"strconv"
	"web"
	"session"
)

type Index struct {}

func (c *Index) Index() string {
	p := PostModelInit()
	results := p.FrontPage()
	
	output := mustache.RenderFile("templates/post.mustache", map[string][]map[string]string {"posts":results})
	return render(output, "")
}

func (c *Index) RSS(ctx *web.Context) string {
	p := PostModelInit()
	results := p.RSS()

	ctx.ContentType("xml")
	return mustache.RenderFile("templates/rss.mustache", map[string][]map[string]string {"posts":results})
}

func (c *Index) NewComment(ctx *web.Context, postId string) {
	p := PostModelInit()
	p.InsertComment(postId,ctx.Params["comment"],ctx.Params["author"])
	ctx.Redirect(302,"/post/"+postId)
}

func (c *Index) ReadPost(ctx *web.Context, postId string) string {
	p := PostModelInit()
	result := p.RenderPost(postId)
	
	viewVars := make(map[string]interface{})
	viewVars["Title"] = result.Title
	viewVars["Content"] = result.Content
	viewVars["Date"] = time.Unix(result.Created, 0).Format(blogConfig.Get("dateFormat"))
	viewVars["Id"] = objectIdHex(result.Id.String())
	// To be used within the {{Comments}} blog
	viewVars["PostId"] = objectIdHex(result.Id.String())
	
	if result.Status == 0 {
		sessionH := session.Start(ctx, h)
		defer sessionH.Save()
		if sessionH.Data["logged"] == nil {
			ctx.Redirect(302, "/")
			return ""
		}
	}
	
	if blogConfig.Get("enableComment") != "" {
		viewVars["EnableComment"] = true
	} else {
		viewVars["EnableComment"] = false
	}
	
	// Render comments
	comments := make([]map[string]string,0)
	for i, v := range result.Comments {
		comments = append(comments, map[string]string{
			"Number": strconv.Itoa(i+1),
			"Date": time.Unix(v.Created, 0).Format(blogConfig.Get("dateFormat")),
			"Id": v.Id[0:9],
			"RealId": v.Id,
			"Content": v.Content,
			"Author": v.Author})
	}
	viewVars["Comments"] = comments
	
	if next, exists := p.GetNextId(objectIdHex(result.Id.String())); exists {
		viewVars["Next"] = next
	}
	if last, exists := p.GetLastId(objectIdHex(result.Id.String())); exists {
		viewVars["Last"] = last
	}
	
	
	sessionH := session.Start(ctx, h)
	defer sessionH.Save()
	viewVars["Admin"] = false
	if sessionH.Data["logged"] != nil {
		viewVars["Admin"] = true
	}
	
	output := mustache.RenderFile("templates/view-post.mustache", viewVars)
	return render(output, result.Title)
}

func (c *Index) ReadPage(pageSlug string) string {
	p := PageModelInit()
	result := p.GetBySlug(pageSlug)
	
	viewVars := make(map[string]string)
	viewVars["Title"] = result.Title
	viewVars["Content"] = result.Content
	viewVars["Date"] = time.Unix(result.Created, 0).Format(blogConfig.Get("dateFormat"))
	
	output := mustache.RenderFile("templates/view-page.mustache", viewVars)
	return render(output, result.Title)
}

func (c *Index) ListPosts(ctx *web.Context) string {
	page := 1
	if temp, exists := ctx.Params["page"]; exists {
		page, _ = strconv.Atoi(temp)
	}
	p := PostModelInit()
	results := p.PostListing(page)
	
	totPages := p.TotalPages()
	
	output := mustache.RenderFile("templates/post-listing.mustache", map[string]interface{} {"Posts": results, "Pagination": pagination(page, totPages)})
	return render(output, "Post Listing")
}