package main

import (
	"launchpad.net/gobson/bson"
	"./mustache"
	"os"
	"time"
	"web"
)

func index(ctx *web.Context) string {
	c := mSession.DB(dbname).C("posts")
	var result *Post
	results := []map[string]string{}
	err := c.Find(nil).Sort(bson.M{"timestamp":-1}).Limit(10).For(&result, func() os.Error {
		t := time.SecondsToLocalTime(result.Timestamp)
		results = append(results, map[string]string {"Title":result.Title, "Content":result.Content, "Date":t.Format("2006 Jan 02 15:04")})
		return nil
	})
	if err != nil {
		panic(err)
	}
	
	output := mustache.RenderFile("templates/post.mustache", map[string][]map[string]string {"posts":results})
	return layout.Render(map[string]string {"Body": output})
}