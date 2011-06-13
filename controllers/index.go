package main

import (
	"./mustache"
	"time"
)

func index() string {
	p := PostModelInit(mSession.DB(config.Get("mongodb")).C("posts"))
	results := p.FrontPage()
	
	output := mustache.RenderFile("templates/post.mustache", map[string][]map[string]string {"posts":results})
	return layout.Render(map[string]string {"Body": output})
}

func readPost(postId string) string {
	p := PostModelInit(mSession.DB(config.Get("mongodb")).C("posts"))
	result := p.Get(postId)
	
	viewVars := make(map[string]string)
	viewVars["Title"] = result.Title
	viewVars["Content"] = result.Content
	viewVars["Date"] = time.SecondsToLocalTime(result.Timestamp).Format("2006 Jan 02 15:04")
	viewVars["Id"] = objectIdHex(result.Id.String())
	
	
	if next, exists := p.GetNextId(objectIdHex(result.Id.String())); exists {
		viewVars["Next"] = next
	}
	if last, exists := p.GetLastId(objectIdHex(result.Id.String())); exists {
		viewVars["Last"] = last
	}
	
	output := mustache.RenderFile("templates/view-post.mustache", viewVars)
	return layout.Render(map[string]interface{} {"Body": output, "Title": map[string]string {"Name": result.Title}})
}

func listPosts() string {
	p := PostModelInit(mSession.DB(config.Get("mongodb")).C("posts"))
	results := p.PostListing()
	
	output := mustache.RenderFile("templates/post-listing.mustache", map[string]interface{} {"Posts": results})
	return layout.Render(map[string]interface{} {"Body": output, "Title": map[string]string {"Name": "Post Listing"}})
}