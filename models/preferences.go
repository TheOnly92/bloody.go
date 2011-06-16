package main

import (
	"launchpad.net/gobson/bson"
	"launchpad.net/mgo"
)

type Preference struct {
	Name		string
	Value		string
}

type PreferenceModel struct {
	c			mgo.Collection
}

func PreferenceInit() *PreferenceModel {
	p := new(PreferenceModel)
	p.c = mSession.DB(config.Get("mongodb")).C("preferences")
	return p
}

func (preference *PreferenceModel) GetByName(name string) *Preference {
	var result *Preference
	err := preference.c.Find(bson.M{"name": name}).One(&result)
	if err == mgo.NotFound {
		// See if there's default value
		value := preference.Get(name)
		result = &Preference{name, value}
	} else if err != nil {
		panic(err)
	}
	return result
}

func (preference *PreferenceModel) Get(name string) string {
	var result *Preference
	err := preference.c.Find(bson.M{"name": name}).One(&result)
	if err == mgo.NotFound {
		// Default values
		switch name {
		case "dateFormat":
			return "2006 Jan 02 15:04"
		case "postsPerPage":
			return "10"
		}
	} else if err != nil {
		panic(err)
	}
	
	return result.Value
}

func (preference *PreferenceModel) Update(name string, value string) {
	result := preference.GetByName(name)
	result.Value = value
	err := preference.c.Update(bson.M{"name":name},result)
	if err == mgo.NotFound {
		err = preference.c.Insert(&Preference{name, value})
		if err != nil {
			panic(err)
		}
	} else if err != nil {
		panic(err)
	}
}