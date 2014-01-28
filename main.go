package main

import (
	"code.google.com/p/gorilla/pat"
	"code.google.com/p/gorilla/sessions"
	"encoding/gob"
	"fmt"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"net/http"
	"crypto/tls"
)
//"code.google.com/p/gorilla/securecookie"

func reverse(name string, things ...interface{}) string {
	//convert the things to strings
	strs := make([]string, len(things))
	for i, th := range things {
		strs[i] = fmt.Sprint(th)
	}
	//grab the route
	u, err := router.GetRoute(name).URL(strs...)
	if err != nil {
		panic(err)
	}
	return u.Path
}

func init() {
	gob.Register(bson.ObjectId(""))
}

var store sessions.Store
var session *mgo.Session
var database string
var router *pat.Router

func main() {
	var err error
	session, err = mgo.Dial("localhost")
	if err != nil {
		panic(err)
	}

	session.SetMode(mgo.Monotonic, true)
	var mydb *mgo.Database
	mydb = session.DB("test")
	loginerr := mydb.Login("youruser","yourpassword")
	if (loginerr != nil) {
		fmt.Printf("couldn't authenticate\n")
		panic(loginerr)
	}


	database = session.DB("test").Name
	fmt.Printf("database name:<<%s>>\n", database)

	//create an index for the username field on the users collection
	if err := session.DB("test").C("users").EnsureIndex(mgo.Index{
		Key:    []string{"username"},
		Unique: true,
	}); err != nil {
		panic(err)
	}


	store = sessions.NewCookieStore([]byte("yoursecretkey"))
	//store = sessions.NewCookieStore([]byte(os.Getenv("KEY")))
	//store = sessions.NewCookieStore(securecookie.GenerateRandomKey(32))

	router = pat.New()
	router.Add("GET", "/images/lmromanunsl10-regular.ttf", handler(getBinaryFile)).Name("getBinaryFile")
	router.Add("GET", "/images/lmromanunsl10-regular.eot", handler(getBinaryFile)).Name("getBinaryFile")
	router.Add("GET", "/images/AdequaTechBusinessCard31May2013.png", handler(getBinaryFile)).Name("getBinaryFile")
	router.Add("GET", "/login", handler(loginForm)).Name("login")
	router.Add("POST", "/login", handler(login))
	router.Add("GET", "/register", handler(registerForm)).Name("register")
	router.Add("POST", "/register", handler(register))
	router.Add("GET", "/upload", handler(uploadForm)).Name("upload")
	router.Add("POST", "/upload", handler(upload))
	router.Add("GET", "/logout", handler(logout)).Name("logout")
	router.Add("POST", "/sign", handler(gostbooksign)).Name("gostbooksign")
	//this needs to be the last route because it is the catch all route
	//when urls can't be found elsewhere
	router.Add("GET", "/", handler(gostbookhello)).Name("gostbookhello")

	// if err = http.ListenAndServeTLS(":5555", "/home/loongson/webServerKeysV2/adequatech.ca-comodoinstantssl-exported-publickey-pem.crt", "/home/loongson/webServerKeysV2/adequatech.ca-comodoinstantssl-exported-privatekey-rsa-ForApache.key", router); err != nil {
	// 	panic(err)
	// }

	//cat comodoSigned/adequatech_ca.crt comodoSigned/COMODOHigh-AssuranceSecureServerCA.crt comodoSigned/AddTrustExternalCARoot.crt > golangCertFile1
	// if err = http.ListenAndServeTLS(":5555", "/home/loongson/webServerKeysV2/golangCertFile1", "/home/loongson/webServerKeysV2/adequatech.ca-comodoinstantssl-exported-privatekey-rsa-ForApache.key", router); err != nil {
	// 	panic(err)
	// }

	//cat comodoSigned/adequatech_ca.crt comodoSigned/AddTrustExternalCARoot.crt comodoSigned/COMODOHigh-AssuranceSecureServerCA.crt > golangCertFile2
	// if err = http.ListenAndServeTLS(":5555", "/home/loongson/webServerKeysV2/golangCertFile2", "/home/loongson/webServerKeysV2/adequatech.ca-comodoinstantssl-exported-privatekey-rsa-ForApache.key", router); err != nil {
	// 	panic(err)
	// }

	myTLSConfig := &tls.Config{
		CipherSuites: []uint16{
			tls.TLS_RSA_WITH_RC4_128_SHA,
			tls.TLS_ECDHE_RSA_WITH_RC4_128_SHA,
			tls.TLS_RSA_WITH_AES_128_CBC_SHA,
			tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA},}
	myTLSConfig.PreferServerCipherSuites = true
	const myWebServerListenAddress = "0.0.0.0:5555"
	myTLSWebServer := &http.Server{Addr: myWebServerListenAddress, TLSConfig: myTLSConfig, Handler: router}
	if err = myTLSWebServer.ListenAndServeTLS("/home/loongson/webServerKeysV2/golangCertFile2", "/home/loongson/webServerKeysV2/adequatech.ca-comodoinstantssl-exported-privatekey-rsa-ForApache.key"); err != nil {
		panic(err)
	}
}
