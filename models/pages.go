package main

import (
	"launchpad.net/gobson/bson"
	"launchpad.net/mgo"
	"time"
	//"math"
)

type Page struct {
	Id			bson.ObjectId "_id,omitempty"
	Title		string
	Slug		string
	Content		string
	Created		int64
	Modified	int64
}

type PageModel struct {
	c			mgo.Collection
}

func PageModelInit() *PageModel {
	p := new(PageModel)
	p.c = mSession.DB(config.Get("mongodb")).C("pages")
	return p
}

func (page *PageModel) Sidebar() []map[string]string {
	var result *Page
	results := []map[string]string{}
	callback := func() error {
		results = append(results, map[string]string {"Title": result.Title, "Slug": result.Slug})
		return nil
	}
	err := page.c.Find(nil).Sort(bson.M{"title":-1}).For(&result, callback)
	if err != nil {
		panic(err)
	}
	return results
}

func (page *PageModel) List() []map[string]string {
	var result *Page
	results := []map[string]string{}
	callback := func() error {
		t := time.Unix(result.Created, 0)
		results = append(results, map[string]string {"Title":result.Title, "Date":t.Format("2006 Jan 02 15:04"), "Id": objectIdHex(result.Id.String())})
		return nil
	}
	var err error
	err = page.c.Find(nil).Sort(bson.M{"timestamp":-1}).For(&result, callback)
	if err != nil {
		panic(err)
	}
	return results
}

func (page *PageModel) Create(title string, content string) {
	t := time.Now()
	slug := toAscii(title)
	if page.GetBySlug(slug) != nil {
		slug += "-2"
	}
	err := page.c.Insert(&Page{"", title, slug, content, t.Unix(), 0})
	if err != nil {
		panic(err)
	}
}

func (page *PageModel) Update(title string, content string, id string) {
	result := page.Get(id)
	result.Title = title
	slug := toAscii(title)
	tmp := page.GetBySlug(slug)
	if tmp != nil {
		if tmp.Id != result.Id {
			slug += "-2"
		}
	}
	result.Slug = slug
	result.Content = content
	result.Modified = time.Now().Unix()
	err := page.c.Update(bson.M{"_id":bson.ObjectIdHex(id)},result)
	if err != nil {
		panic(err)
	}
}

func (page *PageModel) Delete(id string) {
	err := page.c.Remove(bson.M{"_id":bson.ObjectIdHex(id)})
	if err != nil {
		panic(err)
	}
}

func (page *PageModel) Get(id string) *Page {
	var result *Page
	err := page.c.Find(bson.M{"_id": bson.ObjectIdHex(id)}).One(&result)
	if err != nil && err != mgo.NotFound {
		panic(err)
	} else if err == mgo.NotFound {
		return nil
	}
	
	return result
}

func (page *PageModel) GetBySlug(slug string) *Page {
	var result *Page
	err := page.c.Find(bson.M{"slug": slug}).One(&result)
	if err != nil && err != mgo.NotFound {
		panic(err)
	} else if err == mgo.NotFound {
		return nil
	}
	
	return result
}