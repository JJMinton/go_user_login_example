package main


import ("net/http"
        "io/ioutil"
        "log"
        "os"

        "encoding/json"

        //"github.com/gorilla/sessions"
        "github.com/gorilla/securecookie"
        "golang.org/x/oauth2"
        "golang.org/x/oauth2/google"

        "database/sql"
        _ "github.com/mattn/go-sqlite3" //remind me what the underscore does?
        "github.com/gorilla/pat" //alternative to net/http
        )

//Config
//var store *sessions.CookieStore
var store *securecookie.SecureCookie
type Config struct {
    RootURL         string `json:"rootURL"`
    Port            string `json:"port"`
    CookieHashKey   []byte `json:"cookieHashKey"`
    CookieBlockKey  []byte `json:"cookieBlockKey"`
    CookieName      string `json:"cookieName"`
    Signups         bool   `json:"signups"`
    DBType          string `json:"db_type"`
    DBAddress       string `json:"db_address"`
}
var config Config

// Construction config with credentials using init
type Credentials struct {
    Cid     string `json:"cid"`
    Csecret string `json:"csecret"`
    Callback string `json:"callbackURL"`
}
var googleCred Credentials
var googleConf *oauth2.Config

func init() {
    //Generic config
    file, err := ioutil.ReadFile("./server/config.json")
    if err != nil {
        log.Printf("File error: %v\n", err)
        os.Exit(1)
    }
    json.Unmarshal(file, &config)
    log.Printf("Starting server at %s on port %s\n", config.RootURL, config.Port)
    //store = securecookie.New([]byte(config.CookieHashKey), []byte(config.CookieBlockKey))
    store = securecookie.New(config.CookieHashKey, config.CookieBlockKey)

    //Google oauth config
    file, err = ioutil.ReadFile("./server/google_creds.json")
    if err != nil {
        log.Printf("File error: %v\n", err)
        os.Exit(1)
    }
    json.Unmarshal(file, &googleCred)

    googleConf = &oauth2.Config{
        ClientID:       googleCred.Cid,
        ClientSecret:   googleCred.Csecret,
        RedirectURL:    config.RootURL+googleCred.Callback,
        Scopes: []string{
            "email",
        },
        Endpoint: google.Endpoint,
    }
}

//Router
func main() {
    // Front end
    http.HandleFunc("/private/", authenticatePage(http.FileServer(http.Dir("./front-end/dist")).ServeHTTP))
    http.Handle("/", http.FileServer(http.Dir("./front-end/dist/public")))
    // Authentication
    http.HandleFunc("/logout", logoutHandler)
    http.HandleFunc("/auth/google/login", googleLoginHandler)
    http.HandleFunc(googleCred.Callback, googleCallbackHandler)
    // API Endpoints
    r := pat.New()
    r.Get("/endpoints/user", authenticatePage(GetUserDetails))
    r.Post("/endpoints/user", PostUserDetails)
    http.Handle("/endpoints/", r)
    // Start server
    http.ListenAndServe(config.Port, nil)
}


//Logout Handler
func logoutHandler(res http.ResponseWriter, req *http.Request) {
    logout(res, req)
    http.Redirect(res, req, "/index.html", http.StatusSeeOther)
}

//Google login
func googleLoginHandler(res http.ResponseWriter, req *http.Request) {
    const state = "123352lkkjdgdsu"//this should be randomly generated
    //var values = make(url.Values)
    http.Redirect(res, req, googleConf.AuthCodeURL(state), http.StatusSeeOther)
}


type GoogleProfile struct {
    Id              string `json:"id"`
    Email           string `json:"email"`
    Name            string `json:"name"`
    GivenName       string `json:"given_name"`
    FamilyName      string `json:"family_name"`
    Link            string `json:"link"`
    Picture         string `json:"picture"`
    Gender          string `json:"gender"`
}

func googleCallbackHandler(res http.ResponseWriter, req *http.Request) {
    // check for state
    log.Print("Googel call back handler")
    authcode := req.FormValue("code")
    tok, err := googleConf.Exchange(oauth2.NoContext, authcode)
    if err != nil {
        log.Print("err is ", err)
        http.Redirect(res, req, "/private/failed_login.html", http.StatusSeeOther)
    }
    log.Print(string(tok.AccessToken))
    googleProfile, err := getGoogleInfo(tok.AccessToken)
    if err != nil {
        log.Print("Failed to get profile.")
        log.Print(err)
        http.Redirect(res, req, "/private/failed_login.html", http.StatusSeeOther)
        return
    }
    if user_exists, username, err := UserInDB_Google(googleProfile.Id); err != nil {
    } else if user_exists {
        login(res, req, username)
        http.Redirect(res, req, "/private/login_success.html", http.StatusSeeOther)
        return
    } else if config.Signups {
        SetCookie(res, req, Cookie{"false", "google_token", tok.AccessToken})
        log.Print("Not a user, redirected to sign up.")
        http.Redirect(res, req, "/signup.html", http.StatusUnauthorized)
        return
    } else {
        log.Print("Not a user and sign ups not allowed.")
        http.Redirect(res, req, "/failed_login.html", http.StatusUnauthorized)
        return
    }
}

func getGoogleInfo(accessToken string) (GoogleProfile, error) {
    var googleProfile GoogleProfile
    response, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + accessToken)
    if err != nil { return googleProfile, err }
    defer response.Body.Close()
    file, err := ioutil.ReadAll(response.Body)
    if err != nil { return googleProfile, err }
    err = json.Unmarshal(file, &googleProfile)
    if err != nil {
        log.Print("Failed to unmarshal google profile")
        return googleProfile, err
    }
    return googleProfile, err
}

//Authentication/page protection middleware
func authenticatePage(f http.HandlerFunc) http.HandlerFunc {
    return func(res http.ResponseWriter, req *http.Request) {
        log.Print("starting authentication")
        cookie, err := GetCookie(req)
        if err != nil {
            log.Print("Failed to open cookie")
            http.Redirect(res, req, "/failed_login.html", http.StatusInternalServerError)
            return
        }
        if cookie.LoggedIn == "false" || cookie.LoggedIn == "" { //return error/login page
            log.Print("Authentication failed")
            log.Print(cookie.LoggedIn)
            http.Redirect(res, req, "/failed_login.html", http.StatusUnauthorized)
            return
        } else { //return page
            log.Print("Authentication complete")
            f(res, req)
        }
    }
}

//Login/logout cookies
func storeOauth(res http.ResponseWriter, req *http.Request, oauth string, username string) {
    log.Print("saving google in cookies")
    err := SetCookie(res, req, Cookie{"true", oauth, username})
    if err != nil {
        http.Redirect(res, req, "/unknown_error.html", http.StatusInternalServerError)
    }
}

func login(res http.ResponseWriter, req *http.Request, username string) {
    log.Print("saving logged in cookies")
    err := SetCookie(res, req, Cookie{"true", "username", username})
    if err != nil {
        http.Redirect(res, req, "/unknown_error.html", http.StatusInternalServerError)
    }
}

func logout(res http.ResponseWriter, req *http.Request) {
    log.Print("saving logged out cookies")
    err := SetCookie(res, req, Cookie{"false", "", ""})
    if err != nil {
        http.Redirect(res, req, "/unknown_error.html", http.StatusInternalServerError)
    }
}

//Cookie management
type Cookie struct {
    LoggedIn            string
    Oauth               string
    Id                  string
}

func SetCookie(w http.ResponseWriter, r *http.Request, value Cookie) error{
    if encoded, err := store.Encode(config.CookieName, // Encode with something other than cookie name
                                    map[string]string{"loggedin": value.LoggedIn,
                                                      "oauth":    value.Oauth,
                                                      "id":       value.Id}); err == nil {
        cookie := &http.Cookie{
            Name: config.CookieName,
            Value: encoded,
            Path: "/",
        }
        http.SetCookie(w, cookie)
        return nil
    } else {
        log.Print("Cookie write fail")
        log.Print(err)
        return err
    }
}

func GetCookie(r *http.Request) (Cookie, error) {
    value := make(map[string]string)
    if cookie, err := r.Cookie(config.CookieName); err == nil {
        if err = store.Decode(config.CookieName, cookie.Value, &value); err == nil {
            log.Print("Cookies accessed")
            return Cookie{value["loggedin"], value["oauth"], value["id"]}, err
        } else {
            log.Print("failed to decode cookie")
            return Cookie{"", "", ""}, err
        }
    } else {
        log.Print("failed to retrieve cookie")
        log.Print(err)
        return Cookie{"", "", ""}, err
    }
}

// API Endpoints
// User
type User struct {
    UserName    string
    Name        string
}

func DBConn() (*sql.DB, error) {
    db, err := sql.Open(config.DBType, config.DBAddress)
    if err != nil {
        log.Print("Unable to open Db")
        log.Print(err)
        return db, err
    }
    return db, err
}

func AddUserDetails(username string, profile GoogleProfile) (int64, error) {
    var id int64
    db, err := DBConn()
    if err != nil {return id, err}
    defer db.Close()
    add_usr_stmt, err := db.Prepare("INSERT INTO user_details(username, name) VALUES(?, ?)")
    if err != nil {return id, err}
    add_google_stmt, err := db.Prepare("INSERT INTO google_creds(google_id, username, email, picture) VALUES(?, ?, ?, ?)")
    if err != nil {return id, err}

    log.Print("Adding user to database: " + username)
    log.Print(profile)
    dbRes, err := add_usr_stmt.Exec(username, profile.Name)
    if err != nil {return id, err}
    log.Print("Added user details. Adding google profile.")
    _, err = add_google_stmt.Exec(profile.Id, username, profile.Email, profile.Picture)
    if err != nil {return id, err}
    id, err = dbRes.LastInsertId()
    return id, err
}

func PostUserDetails(res http.ResponseWriter, req *http.Request) {
    log.Print("PostUserDetails received")
    if !config.Signups {
        res.WriteHeader(http.StatusUnauthorized)
        res.Write([]byte("sign up not allowed"))
        return
    }
    var m map[string]string //interface{}
    err := json.NewDecoder(req.Body).Decode(&m)
    if err != nil {
        log.Print("Failed to decode body")
        log.Print(err)
        res.WriteHeader(http.StatusInternalServerError)
        res.Write([]byte("error"))
        return
    }

    log.Print("Checking username availability")
    user_exists, username, err := UserInDB(m["username"])
    if err != nil {
        res.WriteHeader(http.StatusInternalServerError)
        res.Write([]byte("error"))
        return
    }
    if user_exists {
        res.WriteHeader(http.StatusInternalServerError)
        res.Write([]byte("username taken"))
        return
    }

    log.Print("Getting cookie.")
    c, err := GetCookie(req)
    if err != nil {
        log.Print(err);
        res.WriteHeader(http.StatusInternalServerError)
        res.Write([]byte("error"))
        return
    }
    if c.Oauth != "google_token" {
        log.Print("No google token in cookie.")
        res.WriteHeader(http.StatusInternalServerError)
        res.Write([]byte("error"))
        return
    }
    
    log.Print("Get google info")
    profile, err := getGoogleInfo(c.Id)
    if err != nil {
        log.Print("Failed to get google profile.")
        log.Print(err)
    }
    id, err := AddUserDetails(username, profile)
    if err != nil {
        log.Print("Error writing to db")
        log.Print(err)
        res.WriteHeader(http.StatusInternalServerError)
        res.Write([]byte("error"))
        return
    }
    js, err := json.Marshal(id)
    if err != nil {
        log.Print("Error marshalling id")
        log.Print(err)
        res.WriteHeader(http.StatusInternalServerError)
        res.Write([]byte("error"))
        return
    }
    res.WriteHeader(http.StatusCreated)
    res.Write(js)
}

func GetUserDetails(res http.ResponseWriter, req *http.Request) {
    db, err := DBConn()
    if err != nil {}
    defer db.Close()

    // Use cookie to set name
    cookie, err := GetCookie(req)
    if err != nil {}
    if cookie.Oauth != "username" {}
    rows, err := db.Query("SELECT username, name FROM user_details WHERE username=?", cookie.Id)
    defer rows.Close()

    var username string
    var name string
    if rows.Next() {
        err = rows.Scan(&username, &name)
        if rows.Next() {
            panic("too many users by that username")
        }
    } else {
    }

    js, err := json.Marshal(User{username, name})
    res.WriteHeader(http.StatusCreated)
    res.Write(js)
}

func UpdateUser(res http.Response, req *http.Request) {

}

func DeleteUser(res http.Response, req *http.Request) {

}


func UserInDB(username string) (bool, string, error) {
    db, err := DBConn()
    if err != nil {}
    defer db.Close()

    var present bool
    err = db.QueryRow("SELECT EXISTS( SELECT 1 FROM user_details WHERE username=?);", username).Scan(&present)
    if err != nil {
        log.Print(err)
        return false, username, err
    }
    return present, username, nil
}


func UserInDB_Google(google_id string) (bool, string, error) {
    db, err := DBConn()
    if err != nil {}
    defer db.Close()

    var username string
    rows, err := db.Query("SELECT username FROM google_creds WHERE google_id=?", google_id)
    defer rows.Close()
    if err != nil {
        return false, "", err
    }
    if rows.Next(){
        rows.Scan(&username)
    }else{
        return false, "", nil
    }
    if rows.Next(){
        panic("Duplicate usernames for google_id")
    }
    if rows.Err() != nil {
        log.Print("Error during reading username from google_id.")
        return false, "", err
    }
    return true, username, nil
}
