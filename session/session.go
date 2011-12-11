package session

import (
	"crypto/sha1"
	"net/http"
	"time"
	"hash"
	"web"
	"launchpad.net/mgo"
	"launchpad.net/gobson/bson"
	"encoding/hex"
	"encoding/json"
	"strconv"
)

type Session struct {
	id		string
	old_id	string
	Data	map[string]interface{}
	handler	*MHandler
}

func Start(ctx *web.Context, handler *MHandler) *Session {
	session := new(Session)
	session.handler = handler
	session.handler.Clean()
	old := false
	if ctx.Cookies != nil {
		if id, exists := ctx.Cookies["bloody_sess"]; exists {
			session.id = id
			old = true
		}
	}
	if !old {
		// Starts new session
		session.generateId()
		session.handler.Store(session.GetID(), nil)
	}
	rt := session.handler.Retrieve(session.GetID())
	json.Unmarshal(rt.SessionData, &session.Data)
	if session.Data == nil {
		t := make(map[string]interface{})
		session.Data = t
	}
	ctx.SetCookie("bloody_sess", session.GetID(), time.Now().Unix() + 3600)
	return session
}

func (session *Session) Save() {
	session.handler.Store(session.id, session.Data)
}

func (session *Session) generateId() string {
	var header = make(http.Header)
	remoteAddr := header.Get("REMOTE_ADDR")
	t := time.Now()
	var h hash.Hash = sha1.New()
	h.Write([]byte(remoteAddr+strconv.FormatInt(t.Unix(),10)))
	session.id = hex.EncodeToString(h.Sum(nil))
	return session.id
}

func (session *Session) GetID() string {
	return session.id
}

func (session *Session) SetID(id string) {
	session.id = id
}

type MHandler struct {
	session		*mgo.Session
}

type sessionRow struct {
	SessionID		string
	ExpirationTS	int64
	SessionData		[]byte
}

func (handler *MHandler) Store(id string, data map[string]interface{}) {
	c := handler.session.DB("bloody").C("sessions")
	t := time.Now()
	b, _ := json.Marshal(data)
	var sav sessionRow
	sav.SessionID = id
	sav.ExpirationTS = t.Unix()+1440
	sav.SessionData = b
	err := c.Update(bson.M{"sessionid": id}, sav)
	if err != nil {
		if err == mgo.NotFound {
			err = c.Insert(&sessionRow{id, t.Unix()+1440, b})
			if err != nil {
				panic(err)
			}
		} else {
			panic(err)
		}
	}
}

func (handler *MHandler) Clean() {
	c := handler.session.DB("bloody").C("sessions")
	t := time.Now()
	err := c.RemoveAll(bson.M{"expirationts": bson.M{"$lt":t.Unix()}})
	if err != nil {
		panic(err)
	}
}

func (handler *MHandler) Remove(id string) {
	c := handler.session.DB("bloody").C("sessions")
	err := c.Remove(bson.M{"sessionid": id})
	if err != nil {
		panic(err)
	}
}

func (handler *MHandler) Retrieve(id string) *sessionRow {
	var rt sessionRow
	c := handler.session.DB("bloody").C("sessions")
	err := c.Find(bson.M{"sessionid": id}).One(&rt)
	if err != nil {
		if err == mgo.NotFound {
			return &sessionRow{id,0,nil}
		}
		panic(err)
	}
	return &rt
}

func (handler *MHandler) SetSession(mSession *mgo.Session) {
	handler.session = mSession
}