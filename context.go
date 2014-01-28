package main

import (
	"code.google.com/p/gorilla/sessions"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"net/http"
	"fmt"
)

type Context struct {
	Database *mgo.Database
	Session  *sessions.Session
	User     *User
}

func (c *Context) Close() {
	c.Database.Session.Close()
}

//C is a convenience function to return a collection from the context database.
func (c *Context) C(name string) *mgo.Collection {
	return c.Database.C(name)
}

func NewContext(req *http.Request) (*Context, error) {
	//get a particular cookie named "adequatechookie" from the session's store
	sess, err := store.Get(req, "adequatechookie")  
	ctx := &Context{
		Database: session.Clone().DB(database),
		Session:  sess,
	}
	if err != nil {
		fmt.Printf("there is no cookie adequatechookie yet\n")
		return ctx, err
	}

	fmt.Printf("there is a cookie adequatechookie.\n")
	//try to fill in the user from the session
	//If a user is actually logged in the cookie will hold the user's objectid.
	if uid, ok := sess.Values["user"].(bson.ObjectId); ok {
		fmt.Printf("there is a user in the cookie.\n")
		err = ctx.C("users").Find(bson.M{"_id": uid}).One(&ctx.User)
	}

	return ctx, err
}
