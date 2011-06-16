package main

import (
	"./mustache"
	"time"
	"strconv"
	"web"
	"./session"
)

func index() string {
	p := PostModelInit()
	results := p.FrontPage()
	
	output := mustache.RenderFile("templates/post.mustache", map[string][]map[string]string {"posts":results})
	return render(output, "")
}

func readPost(ctx *web.Context, postId string) string {
	p := PostModelInit()
	result := p.Get(postId)
	
	viewVars := make(map[string]string)
	viewVars["Title"] = result.Title
	viewVars["Content"] = result.Content
	viewVars["Date"] = time.SecondsToLocalTime(result.Created).Format(blogConfig.Get("dateFormat"))
	viewVars["Id"] = objectIdHex(result.Id.String())
	
	if result.Status == 0 {
		sessionH := session.Start(ctx, h)
		defer sessionH.Save()
		if sessionH.Data["logged"] == nil {
			ctx.Redirect(302, "/")
			return ""
		}
	}
	
	
	if next, exists := p.GetNextId(objectIdHex(result.Id.String())); exists {
		viewVars["Next"] = next
	}
	if last, exists := p.GetLastId(objectIdHex(result.Id.String())); exists {
		viewVars["Last"] = last
	}
	
	output := mustache.RenderFile("templates/view-post.mustache", viewVars)
	return render(output, result.Title)
}

func readPage(pageSlug string) string {
	p := PageModelInit()
	result := p.GetBySlug(pageSlug)
	
	viewVars := make(map[string]string)
	viewVars["Title"] = result.Title
	viewVars["Content"] = result.Content
	viewVars["Date"] = time.SecondsToLocalTime(result.Created).Format(blogConfig.Get("dateFormat"))
	
	output := mustache.RenderFile("templates/view-page.mustache", viewVars)
	return render(output, result.Title)
}

func listPosts(ctx *web.Context) string {
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