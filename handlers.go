package main

import (
	"errors"
	"labix.org/v2/mgo/bson"
	"net/http"
	"fmt"
	"regexp"
	"io"
	"os"
	"path/filepath"
)

func getBinaryFile(w http.ResponseWriter, req *http.Request, ctx *Context) error {
	//this strategy is deliberately not dynamic
	//for security reasons we don't want strings to directly access files
	//on the web server to prevent
	//vulnerability from relative paths or obfuscated exploit code.
	fmt.Printf("getBinaryFilethe requested path: <<%s>>", req.URL.Path)
	binaryFileName := req.URL.Path
	switch binaryFileName { 
	case "/images/lmromanunsl10-regular.eot": 
		w.Header().Set("Content-Type", "application/vnd.ms-fontobject")
		http.ServeFile(w, req, "images/lmromanunsl10-regular.eot")
	case "/images/lmromanunsl10-regular.ttf": 
		w.Header().Set("Content-Type", "font/ttf")
		http.ServeFile(w, req, "images/lmromanunsl10-regular.ttf")
	case "/images/AdequaTechBusinessCard31May2013.png": 
		w.Header().Set("Content-Type", "image/png")
		http.ServeFile(w, req, "images/AdequaTechBusinessCard31May2013.png")
	}

	// image/gif 
	// image/jpeg
	// image/pjpeg
	// image/png
	// image/svg+xml
	// image/tiff
	// application/vnd.ms-fontobject //.eot
	// application/octet-stream //.otf .ttf
	// font/ttf //.ttf
	// font/otf //.otf
	// application/x-woff //.woff
	//http.Redirect(w, req, reverse("gostbookhello"), http.StatusSeeOther)
	return nil
}

func gostbookhello(w http.ResponseWriter, req *http.Request, ctx *Context) (err error) {
	fmt.Printf("gostbookhellothe <<%s>> requested path: <<%s>>", req.RemoteAddr, req.URL.Path)
	//set up the collection and query
	coll := ctx.C("entries")
	query := coll.Find(nil).Sort("-timestamp")

	//execute the query
	//TODO: add pagination :)
	var entries []Entry
	if err = query.All(&entries); err != nil {
		return
	}

	//execute the template
	return T("gostbookhello.html").Execute(w, map[string]interface{}{
		"entries": entries,
		"ctx":     ctx,
	})
}

func gostbooksign(w http.ResponseWriter, req *http.Request, ctx *Context) (err error) {
	fmt.Printf("gostbooksignthe <<%s>> requested path: <<%s>>", req.RemoteAddr, req.URL.Path)
	//we need a user to sign to
	if ctx.User == nil {
		err = errors.New("Can't sign without being logged in")
		return
	}

	entry := NewEntry()
	entry.Name = ctx.User.Username
	entry.Message = req.FormValue("message")

	if entry.Message == "" {
		entry.Message = "Some dummy who forgot a message."
	}

	coll := ctx.C("entries")
	if err = coll.Insert(entry); err != nil {
		return
	}

	//ignore errors: it's ok if the post count is wrong. we can always look at
	//the entries table to fix.
	ctx.C("users").Update(bson.M{"_id": ctx.User.ID}, bson.M{
		"$inc": bson.M{"posts": 1},
	})

	http.Redirect(w, req, reverse("gostbookhello"), http.StatusSeeOther)
	return
}

func loginForm(w http.ResponseWriter, req *http.Request, ctx *Context) (err error) {
	fmt.Printf("loginFormthe <<%s>> requested path: <<%s>>", req.RemoteAddr, req.URL.Path)
	return T("login.html").Execute(w, map[string]interface{}{
		"ctx": ctx,
	})
}

func login(w http.ResponseWriter, req *http.Request, ctx *Context) error {
	fmt.Printf("loginthe <<%s>> requested path: <<%s>>", req.RemoteAddr, req.URL.Path)
	username, password := req.FormValue("username"), req.FormValue("password")

	user, e := Login(ctx, username, password)
	if e != nil {
		ctx.Session.AddFlash("Invalid Username/Password")
		return loginForm(w, req, ctx)
	}

	//store the user id in the values and redirect to index
	ctx.Session.Values["user"] = user.ID
	http.Redirect(w, req, reverse("gostbookhello"), http.StatusSeeOther)
	return nil
}

func logout(w http.ResponseWriter, req *http.Request, ctx *Context) (err error) {
	if ctx.User == nil {
		err = errors.New("You may not logout without being logged in.  Please log in.")
		return err
	}
	fmt.Printf("logoutthe <<%s>> requested path: <<%s>>", req.RemoteAddr, req.URL.Path)
	delete(ctx.Session.Values, "user")
	http.Redirect(w, req, reverse("gostbookhello"), http.StatusSeeOther)
	return nil
}

func registerForm(w http.ResponseWriter, req *http.Request, ctx *Context) (err error) {
	fmt.Printf("registerFormthe <<%s>> requested path: <<%s>>", req.RemoteAddr, req.URL.Path)
	return T("register.html").Execute(w, map[string]interface{}{
		"ctx": ctx,
	})
}

func register(w http.ResponseWriter, req *http.Request, ctx *Context) error {
	fmt.Printf("registerthe <<%s>> requested path: <<%s>>", req.RemoteAddr, req.URL.Path)
	username := req.FormValue("username")
	password := req.FormValue("password")
	password2 := req.FormValue("password2")

	//double-check password was typed-in correctly twice
	if (password != password2) {
		ctx.Session.AddFlash("Register password must be typed in correctly twice.")
		return registerForm(w, req, ctx)
	}

	if ( len(password) < 12 ) {
		ctx.Session.AddFlash("Register password must be at least 12 characters long with lowercase, uppercase, digits and punctuation marks.")
		return registerForm(w, req, ctx)
	}

	//enforce password contains lower, upper, digit, punct
	lowerRx := regexp.MustCompile("[[:lower:]]")
	upperRx := regexp.MustCompile("[[:upper:]]")
	digitRx := regexp.MustCompile("[[:digit:]]")
	punctRx := regexp.MustCompile("[[:punct:]]")

	if lowerMatches := lowerRx.FindAllStringSubmatch(password, -1); lowerMatches == nil {
		ctx.Session.AddFlash("Register password must have lowercase letters.")
		return registerForm(w, req, ctx)
	}
	if upperMatches := upperRx.FindAllStringSubmatch(password, -1); upperMatches == nil {
		ctx.Session.AddFlash("Register password must have uppercase letters.")
		return registerForm(w, req, ctx)
	}
	if digitMatches := digitRx.FindAllStringSubmatch(password, -1); digitMatches == nil {
		ctx.Session.AddFlash("Register password must have digits.")
		return registerForm(w, req, ctx)
	}
	if punctMatches := punctRx.FindAllStringSubmatch(password, -1); punctMatches == nil {
		ctx.Session.AddFlash("Register password must have punctuation marks.")
		return registerForm(w, req, ctx)
	}

	u := &User{
		Username: username,
		ID:       bson.NewObjectId(),
	}
	u.SetPassword(password)

	if err := ctx.C("users").Insert(u); err != nil {
		ctx.Session.AddFlash("Problem registering user.")
		return registerForm(w, req, ctx)
	}

	//store the user id in the values and redirect to index
	ctx.Session.Values["user"] = u.ID
	http.Redirect(w, req, reverse("gostbookhello"), http.StatusSeeOther)
	return nil
}

func uploadForm(w http.ResponseWriter, req *http.Request, ctx *Context) (err error) {
	fmt.Printf("uploadForm the <<%s>> requested path: <<%s>>", req.RemoteAddr, req.URL.Path)
	return T("upload.html").Execute(w, map[string]interface{}{
		"ctx": ctx,
	})
}

func upload(w http.ResponseWriter, req *http.Request, ctx *Context) (err error) {
	//make sure the user is logged in
	if ctx.User == nil {
		err = errors.New("You may not upload files without being logged in.  Please log in.")
		return
	}

	err2 := req.ParseMultipartForm(10000000) //Parse using 10MB memory chunks
	if err2 != nil {
		ctx.Session.AddFlash("upload files error: %s", err2.Error())
		return
	}

	//get a ref to the parsed multipart form
	m := req.MultipartForm

	//get the *fileheaders
	files := m.File["myfiles"]
	for i, _ := range files {
		// for ensuring no direct access to local file system
		// if stripped filename isn't valid, report it and do nothing with this upload callback.
		myUncleanfullpath := files[i].Filename
		fmt.Printf("unclean files[i].Filename: %s\n", files[i].Filename)
		myCleanedPath := filepath.Clean(myUncleanfullpath) //remove . and .. characters
		fmt.Printf("myCleanedPath:%s\n", myCleanedPath)
		myBasename := filepath.Base(myCleanedPath) //just get basename of file, but it can return a "." character.
		fmt.Printf("myBasename:%s\n", myBasename)
		if (myBasename == ".") {
			ctx.Session.AddFlash("upload files error has no valid filename", err2.Error())
			return			
		}
		myFilenameExtension := filepath.Ext(myBasename)
		fmt.Printf("Ext:%s\n", myFilenameExtension)

		//for each fileheader, get a handle to the actual file
		file, err2 := files[i].Open()
		defer file.Close()
		if err2 != nil {
			ctx.Session.AddFlash("upload files open error: %s", err2.Error())
			return
		}
		//create destination file making sure the path is writeable.
		dst, err2 := os.Create("/home/loongson/Code/go/src/github.com/omac777/t1/uploaded/" + myBasename)
		defer dst.Close()
		if err2 != nil {
			ctx.Session.AddFlash("upload files destination error: %s", err2.Error())
			return
		}
		//copy the uploaded file to the destination file
		fmt.Printf("uploading %s\n", myBasename)
		if _, err2 := io.Copy(dst, file); err2 != nil {
			ctx.Session.AddFlash("upload files write to destination error: %s", err2.Error())
			return
		}
		fmt.Printf("upload successful %s\n", myBasename)

		// validate all the content to match the extension to ensure no file content pretending to be another content type with false filename extensions.
		// if file contents does not match file extension, report it and do nothing with this upload callback.


		//later on transfer all this file into the mongodb gridfs system for better scalability and performance.
		//this web server works through i2p when assigned to an i2p specific port
		//this means i2p is just as powerful as tor.
	}
	//display success message.
	ctx.Session.AddFlash("Upload successful")
	http.Redirect(w, req, reverse("upload"), http.StatusSeeOther)
	return

	// this way shows the progress as you upload parts of a file
	// and the different mime parts...
	// mr, err := req.MultipartReader()
	//     if err != nil {
		//         return
	//     }
	//     length := req.ContentLength
	//     for {

	//         part, err := mr.NextPart()
	//         if err == io.EOF {
	//             break
	//         }
	//         var read int64
	//         var p float32
	//         dst, err := os.OpenFile("dstfile", os.O_WRONLY|os.O_CREATE, 0644)
	//         if err != nil {
	//             return
	//         }
	//         for {
	//             buffer := make([]byte, 100000)
	//             cBytes, err := part.Read(buffer)
	//             if err == io.EOF {
	//                 break
	//             }
	//             read = read + int64(cBytes)
	//             //fmt.Printf("read: %v \n",read )
	//             p = float32(read) / float32(length) *100
	//             fmt.Printf("progress: %v \n",p )
	//             dst.Write(buffer)
	//         }
	//     }


}
