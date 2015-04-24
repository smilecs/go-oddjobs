package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/sessions"
	//"strings"
)

//HomeHandler serves the home/search page to the user
func HomeHandler(w http.ResponseWriter, r *http.Request) {
	type datastruct struct {
		User  LoginDataStruct
		FBURL string
	}

	data := datastruct{
		User:  LoginData(r),
		FBURL: FBURL,
	}

	renderTemplate(w, "index.html", data)
}

//SearchHandler serves the search results page based on a search query from the
//index page or any search box
func SearchHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	location := r.FormValue("l")
	query := r.FormValue("q")
	d, p, err := Search(location, query, 1, 20)
	if err != nil {
		checkFmt(err)
	}
	type datastruct struct {
		User  LoginDataStruct
		FBURL string
		Page  Page
		Data  []Skill
		L     string
		Q     string
	}

	data := datastruct{
		User:  LoginData(r),
		FBURL: FBURL,
		Data:  d,
		Page:  p,
		L:     location,
		Q:     query,
	}
	renderTemplate(w, "search-results.html", data)
}

//SingleHandlerWeb serves the search results page based on a search query from the
//index page or any search box
func SingleHandlerWeb(w http.ResponseWriter, r *http.Request) {

	URL := strings.Split(r.URL.Path, "/")
	location := URL[2]
	slug := URL[3]
	log.Println(slug)

	if r.Method == "GET" {

		skill, err := GetSkillBySlug(slug, location)
		checkFmt(err)

		skillid := SlugtoID(slug)

		log.Println(skillid)

		reviews, err := GetReviews(skillid)
		checkFmt(err)

		log.Println(reviews)
		
		
		
		var zzz []Review
		
		for _,rr := range reviews{
		  uu, err := GetProfile(rr.Id)
		  checkFmt(err)
		  rr.User = uu
		  
		  zzz = append(zzz, rr)
		}
		
		type datastruct struct {
			User    LoginDataStruct
			FBURL   string
			Data    Skill
			Reviews []Review
		}

		data := datastruct{
			User:    LoginData(r),
			FBURL:   FBURL,
			Data:    skill,
			Reviews: zzz,
		}

		renderTemplate(w, "single.html", data)

	} else if r.Method == "POST" {
		log.Println("POSTED review")
		r.ParseForm()

		rate := r.FormValue("rating")
		log.Println(rate)

		review := r.FormValue("description")
		log.Println(review)

		session, err := store.Get(r, "user")
		checkFmt(err)
		
		s, err := strconv.Atoi(rate)
		
		checkFmt(err)
		
		id := session.Values["id"].(string)
		pid := SlugtoID(slug)
		
		
		rr := Review{
			Comment: review,
			Rating:  s,
			Id:      id,
			PostID:  pid,
		}

		log.Println(rr)

		err = AddReview(&rr)
		checkFmt(err)
		http.Redirect(w, r, r.URL.String(), http.StatusFound)
	}
}

//ProfileHandler might be remove later, its just to test redirection and profile
//data collection after login
func ProfileHandler(w http.ResponseWriter, r *http.Request) {
	type datastruct struct {
		User  LoginDataStruct
		FBURL string
	}

	data := datastruct{
		User:  LoginData(r),
		FBURL: FBURL,
	}
	if r.Method == "GET" {
		renderTemplate(w, "profile.html", data)
	} else if r.Method == "POST" {
		fmt.Println("POST request logged")
	}
}

//ProfileEditHandler for now just logs the json value sent by the web client for
//debugging purposes
func ProfileEditHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept")
	fmt.Println(r.Method)
	session, err := store.Get(r, "user")
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(session.Values["id"])
	fmt.Println(session.Values["email"])

	id := session.Values["id"].(string)

	if r.Method == "GET" {
		fmt.Println("Get request")
		fmt.Println(id)
		user, err := GetProfile(id)
		checkFmt(err)
		x, err := json.Marshal(user)
		checkFmt(err)
		fmt.Println("Profile GET user data")
		fmt.Println(user)
		w.Header().Set("Content-Type", "application/json")
		_, err = w.Write(x)

		checkFmt(err)

	} else if r.Method == "POST" {
		hah, err := ioutil.ReadAll(r.Body)
		checkFmt(err)

		fmt.Println(string(hah))
		user := User{}

		err = json.Unmarshal(hah, &user)

		checkFmt(err)
		fmt.Println(user)

		session.Values["email"] = user.Email
		session.Values["name"] = user.Name

		err = session.Save(r, w)
		checkFmt(err)
		err = UpdateUser(&user, id)
		checkFmt(err)
	}
}

//Login ish
func Login(w http.ResponseWriter, r *http.Request) {
	session, err := store.Get(r, "current")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(r.URL.String())

	session.Values["url"] = r.URL.String()

	http.Redirect(w, r, FBURL, http.StatusFound)
}

//Logout dsfsdgs
func Logout(w http.ResponseWriter, r *http.Request) {
	session, err := store.Get(r, "user")
	checkFmt(err)
	session.Options = &sessions.Options{MaxAge: -1, Path: "/"}
	session.Save(r, w)
	http.Redirect(w, r, "/", http.StatusFound)
}

//SkillsHandler would return list of skills via json, and suport editing and
//addition of new skills
func SkillsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept")
	session, err := store.Get(r, "user")
	checkFmt(err)
	fmt.Println(session.Values["id"])

	id := session.Values["id"].(string)

	if r.Method == "GET" {
		fmt.Println("get request")

		skills := []Skill{}

		skills, err := GetSkills(id)
		checkFmt(err)

		x, err := json.Marshal(skills)
		fmt.Print(string(x))
		if err != nil {
			fmt.Println(err)
		}
		w.Header().Set("Content-Type", "application/json")
		_, err = w.Write(x)

		if err != nil {
			fmt.Println(err)
		}

	} else if r.Method == "POST" {
		fmt.Println("post request")
		hah, err := ioutil.ReadAll(r.Body)
		checkFmt(err)

		skill := Skill{}

		err = json.Unmarshal(hah, &skill)
		//fmt.Println(hah)
		//fmt.Println(skill)
		checkFmt(err)

		skill.UserID = id

		err = AddSkill(&skill)
		checkFmt(err)
		http.Redirect(w, r, r.URL.String(), 301)

		/*
			x, err := json.Marshal(skills)
			fmt.Print(string(x))
			if err != nil {
				fmt.Println(err)
			}
			w.Header().Set("Content-Type", "application/json")
			_, err = w.Write(x)

			if err != nil {
				fmt.Println(err)
			}

		*/
	}

}
