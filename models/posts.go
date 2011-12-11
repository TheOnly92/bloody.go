package main

import (
	"launchpad.net/gobson/bson"
	"launchpad.net/mgo"
	"time"
	"hash"
	"crypto/sha1"
	"encoding/hex"
	"math"
	"strconv"
	"github.com/russross/blackfriday"
)

type Post struct {
	Id			bson.ObjectId "_id,omitempty"
	Title		string
	Content		string
	Created		int64 "timestamp"
	Modified	int64
	Status		uint64
	Type		uint
	Comments	[]Comment
}

type Comment struct {
	Id			string
	Content		string
	Author		string
	Created		int64
}

type PostModel struct {
	c			mgo.Collection
	extensions	int
	html_flags	int
}

func PostModelInit() *PostModel {
	p := new(PostModel)
	p.c = mSession.DB(config.Get("mongodb")).C("posts")
	p.extensions |= blackfriday.EXTENSION_NO_INTRA_EMPHASIS
	p.extensions |= blackfriday.EXTENSION_TABLES
	p.extensions |= blackfriday.EXTENSION_FENCED_CODE
	p.extensions |= blackfriday.EXTENSION_AUTOLINK
	p.extensions |= blackfriday.EXTENSION_STRIKETHROUGH
	p.extensions |= blackfriday.EXTENSION_SPACE_HEADERS
	p.extensions |= blackfriday.EXTENSION_HARD_LINE_BREAK
	p.html_flags |= blackfriday.HTML_USE_XHTML
	p.html_flags |= blackfriday.HTML_USE_SMARTYPANTS
	p.html_flags |= blackfriday.HTML_SMARTYPANTS_FRACTIONS
	p.html_flags |= blackfriday.HTML_SMARTYPANTS_LATEX_DASHES
	return p
}

func (post *PostModel) FrontPage() []map[string]string {
	var result *Post
	results := []map[string]string{}
	posts, _ := strconv.Atoi(blogConfig.Get("postsPerPage"))
	err := post.c.Find(bson.M{"status":1}).Sort(bson.M{"timestamp":-1}).Limit(posts).For(&result, func() error {
		t := time.Unix(result.Created, 0)
		if result.Type == 1 {
			renderer := blackfriday.HtmlRenderer(post.html_flags,"","")
			result.Content = string(blackfriday.Markdown([]byte(result.Content), renderer, post.extensions))
		}
		results = append(results, map[string]string {"Title":result.Title, "Content":result.Content, "Date":t.Format(blogConfig.Get("dateFormat")), "Id": objectIdHex(result.Id.String())})
		return nil
	})
	if err != nil {
		panic(err)
	}
	
	return results
}

func (post *PostModel) RSS() []map[string]string {
	var result *Post
	results := []map[string]string{}
	err := post.c.Find(bson.M{"status":1}).Sort(bson.M{"timestamp":-1}).Limit(20).For(&result, func() error {
		t := time.Unix(result.Created, 0)
		if result.Type == 1 {
			renderer := blackfriday.HtmlRenderer(post.html_flags,"","")
			result.Content = string(blackfriday.Markdown([]byte(result.Content), renderer, post.extensions))
		}
		results = append(results, map[string]string {"Title":result.Title, "Content":result.Content, "Date":t.Format(time.RFC1123), "Id": objectIdHex(result.Id.String())})
		return nil
	})
	if err != nil {
		panic(err)
	}
	
	return results
}

func (post *PostModel) RenderPost(postId string) *Post {
	result := post.Get(postId)
	renderer := blackfriday.HtmlRenderer(post.html_flags,"","")
	if result.Type == 1 {
		result.Content = string(blackfriday.Markdown([]byte(result.Content), renderer, post.extensions))
	}
	
	// Also render all blog comments as markdown
	for i, v := range result.Comments {
		result.Comments[i].Content = string(blackfriday.Markdown([]byte(v.Content), renderer, post.extensions))
	}
	return result
}

func (post *PostModel) Get(postId string) *Post {
	var result *Post
	err := post.c.Find(bson.M{"_id": bson.ObjectIdHex(postId)}).One(&result)
	if err != nil {
		panic(err)
	}
	
	return result
}

func (post *PostModel) GetNextId(postId string) (string, bool) {
	var next *Post
	result := post.Get(postId)
	err := post.c.Find(bson.M{"status":1, "timestamp": bson.M{"$gt":result.Created}}).Sort(bson.M{"timestamp":1}).One(&next)
	if err != nil && err != mgo.NotFound {
		panic(err)
	}
	
	if err == mgo.NotFound {
		return "", false
	}
	return objectIdHex(next.Id.String()), true
}

func (post *PostModel) GetLastId(postId string) (string, bool) {
	var last *Post
	result := post.Get(postId)
	err := post.c.Find(bson.M{"status":1, "timestamp": bson.M{"$lt":result.Created}}).Sort(bson.M{"timestamp":-1}).One(&last)
	if err != nil && err != mgo.NotFound {
		panic(err)
	}
	
	if err == mgo.NotFound {
		return "", false
	}
	return objectIdHex(last.Id.String()), true
}

func (post *PostModel) TotalPages() int {
	total, err := post.c.Find(bson.M{"status":1}).Count()
	if err != nil {
		panic(err)
	}
	posts, _ := strconv.Atoi(blogConfig.Get("postsPerPage"))
	pages := float64(total) / float64(posts)
	return int(math.Ceil(pages))
}

func (post *PostModel) PostListing(page int) []map[string]string {
	var result *Post
	results := []map[string]string{}
	callback := func() error {
		t := time.Unix(result.Created, 0)
		p := map[string]string {"Title":result.Title, "Date":t.Format(blogConfig.Get("dateFormat")), "Id": objectIdHex(result.Id.String())}
		if (result.Status == 0) {
			p["Draft"] = "1"
		}
		results = append(results, p)
		return nil
	}
	var err error
	if page == 0 {
		err = post.c.Find(nil).Sort(bson.M{"timestamp":-1}).For(&result, callback)
	} else {
		posts, _ := strconv.Atoi(blogConfig.Get("postsPerPage"))
		err = post.c.Find(bson.M{"status":1}).Sort(bson.M{"timestamp":-1}).Skip(posts * (page - 1)).Limit(posts).For(&result, callback)
	}
	if err != nil {
		panic(err)
	}
	return results
}

func (post *PostModel) Create(title string, content string, status string, markdown uint) {
	t := time.Now()
	tmp, _ := strconv.ParseUint(status, 10, 64)
	err := post.c.Insert(&Post{"", title, content, t.Unix(), 0, tmp, markdown, make([]Comment,0)})
	if err != nil {
		panic(err)
	}
}

func (post *PostModel) InsertComment(postId string, content string, author string) {
	result := post.Get(postId)
	if author == "" {
		author = "Anonymous"
	}
	// Generate ID
	t := time.Now()
	var h hash.Hash = sha1.New()
	h.Write([]byte(author+strconv.FormatInt(t.Unix(),10)))
	id := hex.EncodeToString(h.Sum(nil))
	comment := Comment{id, content, author, t.Unix()}
	result.Comments = append(result.Comments, comment)
	err := post.c.Update(bson.M{"_id":bson.ObjectIdHex(postId)},result)
	if err != nil {
		panic(err)
	}
}

func (post *PostModel) DeleteComment(postId string, id string) {
	result := post.Get(postId)
	
	// Copy everything from the old one and skip those that we want to delete
	newComments := make([]Comment,0)
	for _, comment := range result.Comments {
		if comment.Id != id {
			newComments = append(newComments, comment)
		}
	}
	result.Comments = newComments
	
	// Save result
	err := post.c.Update(bson.M{"_id":bson.ObjectIdHex(postId)},result)
	if err != nil {
		panic(err)
	}
}

func (post *PostModel) Update(title string, content string, status string, postId string) {
	result := post.Get(postId)
	result.Title = title
	result.Content = content
	result.Modified = time.Now().Unix()
	result.Status, _ = strconv.ParseUint(status,0,64)
	err := post.c.Update(bson.M{"_id":bson.ObjectIdHex(postId)},result)
	if err != nil {
		panic(err)
	}
}

func (post *PostModel) Publish(postId string) {
	result := post.Get(postId)
	result.Status = 1
	err := post.c.Update(bson.M{"_id":bson.ObjectIdHex(postId)},result)
	if err != nil {
		panic(err)
	}
}

func (post *PostModel) Delete(postId string) {
	err := post.c.Remove(bson.M{"_id":bson.ObjectIdHex(postId)})
	if err != nil {
		panic(err)
	}
}

func (post *PostModel) DeleteBulk(postIds []string) {
	for _, v := range postIds {
		post.Delete(v)
	}
}

func (post *PostModel) PublishBulk(postIds []string) {
	for _, v := range postIds {
		post.Publish(v)
	}
}