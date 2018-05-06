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
    http.HandleFunc("/endpoints/server_name", authenticatePage(dataEndpoint))
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

type Profile struct {
    Id              string `json:"id"`
    Name            string `json:"name"`
    PreferredEmail  string `json:"preferred_email"`
    PhotoUrl        string `json:"photo_url"`
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
    authcode := req.FormValue("code")
    tok, err := googleConf.Exchange(oauth2.NoContext, authcode)
    if err != nil {
        log.Fatal("err is ", err)
    }
    log.Print(string(tok.AccessToken))
    response, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + tok.AccessToken)
    defer response.Body.Close()
    file, err := ioutil.ReadAll(response.Body)
    if err != nil {
      log.Fatal("failed to read google respones body")
    }
    var googleProfile GoogleProfile
    json.Unmarshal(file, &googleProfile)
    log.Print(string(file))
    //TODO: do a check and login here.
    log.Print("logging in from google callback handler")
    login(res, req, Profile{Id: googleProfile.Id,
                            Name: googleProfile.Name,
                            PreferredEmail: googleProfile.Email,
                            PhotoUrl: googleProfile.Picture})
    http.Redirect(res, req, "/private/login_success.html", http.StatusSeeOther)
}



//Authentication/page protection middleware
func authenticatePage(f http.HandlerFunc) http.HandlerFunc {
    return func(res http.ResponseWriter, req *http.Request) {
        log.Print("starting authentication")
        cookie, err := GetCookie(req)
        if err != nil {
            log.Fatal("Failed to open cookie")
        }
        loggedin := string(cookie["loggedin"])
        if loggedin == "false" || loggedin == "" { //return error/login page
            log.Print("authentication failed")
            log.Print(loggedin)
            http.Redirect(res, req, "/failed_login.html", http.StatusSeeOther)
        } else { //return page
            log.Print("authentication complete")
            f(res, req)
        }
    }
}

//Login/logout cookies
func login(res http.ResponseWriter, req *http.Request, profile Profile) {
    log.Print("saving logged in cookies")
    err := SetCookie(res, req, map[string]string{"loggedin": "true",
                                                 "name": profile.Name,
                                                 "id": profile.PreferredEmail,
                                                 "photo": profile.PhotoUrl})
    if err != nil {
        log.Fatal("Failed to save cookie")
    }
}

func logout(res http.ResponseWriter, req *http.Request) {
    log.Print("saving logged out cookies")
    err := SetCookie(res, req, map[string]string{"loggedin": "false",})
    if err != nil {
        log.Fatal("Failed to save cookie")
    }
}

//Cookie management
func SetCookie(w http.ResponseWriter, r *http.Request, value map[string]string) error{
    if encoded, err := store.Encode(config.CookieName, value); err == nil {
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

func GetCookie(r *http.Request) (map[string]string, error) {
    if cookie, err := r.Cookie(config.CookieName); err == nil {
        value := make(map[string]string)
        if err = store.Decode(config.CookieName, cookie.Value, &value); err == nil {
            log.Print("Cookies accessed")
            return value, err
        } else {
            log.Print("failed to decode cookie")
            return make(map[string]string), err
        }
    } else {
        log.Print("failed to retrieve cookie")
        log.Print(err)
        return make(map[string]string), err
    }
}

//API Endpoints
func dataEndpoint(res http.ResponseWriter, req *http.Request) {
    res.Header().Set("Content-Type", "application/json")
    cookie, err := GetCookie(req)
    if err != nil {
        log.Fatal("Failed to open cookie")
    }
    res.WriteHeader(http.StatusOK)
    if cookieJson, err := json.Marshal(cookie); err == nil {
        res.Write(cookieJson)
    } else {
        log.Fatal("failed to marshal cookie")
    }
}
