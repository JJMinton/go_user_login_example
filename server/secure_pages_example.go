package main


import ("net/http"
        "io"
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
    http.HandleFunc("/", rootHandler)
    http.HandleFunc("/auth/google/login", googleLoginHandler)
    http.HandleFunc(googleCred.Callback, googleCallbackHandler)
    http.HandleFunc("/logout", logoutHandler)
    http.HandleFunc("/protectedpage", authenticatePage(protectedPageHandler))
    http.ListenAndServe(config.Port, nil)
}


//Handlers
func rootHandler(res http.ResponseWriter, req *http.Request) {
    io.WriteString(res, `
<!DOCTYPE html>
<html>
  <head></head>
  <body>
    <p><a href="/auth/google/login">LOGIN</a></p>
    <p><a href="/logout">LOGOUT</a></p>
    <p><a href="/protectedpage">Test login</a></p>
  </body>
</html>`)
}

func failedLoginHandler(res http.ResponseWriter, req *http.Request) {
    io.WriteString(res, `
<!DOCTYPE html>
<html>
  <head></head>
  <body>
    <p>Failed login, try again.</p>
    <a href="/">Login Page</a>
  </body>
</html>`)
}

func logoutHandler(res http.ResponseWriter, req *http.Request) {
    logout(res, req)
    http.Redirect(res, req, "/", http.StatusSeeOther)
}

func protectedPageHandler(res http.ResponseWriter, req *http.Request) {
    io.WriteString(res, `
<!DOCTYPE html>
<html>
  <head></head>
  <body>
    <p>This page shouldn't be accessible without logging in</p>
    <a href="/logout">Logout</a>
  </body>
</html>`)

}

//Google login
func googleLoginHandler(res http.ResponseWriter, req *http.Request) {
    const state = "123352lkkjdgdsu"//this should be randomly generated
    //var values = make(url.Values)
    http.Redirect(res, req, googleConf.AuthCodeURL(state), http.StatusSeeOther)
}

func googleCallbackHandler(w http.ResponseWriter, req *http.Request) {
    authcode := req.FormValue("code")
    tok, err := googleConf.Exchange(oauth2.NoContext, authcode)
    if err != nil {
        log.Fatal("err is ", err)
    }
    log.Print(string(tok.AccessToken))
    response, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + tok.AccessToken)
    defer response.Body.Close()
    contents, err := ioutil.ReadAll(response.Body)
    if err != nil {
      log.Fatal("failed to read google respones body")
    }
    log.Print(string(contents))
    //TODO: do a check and login here.
    log.Print("logging in from google callback handler")
    login(w, req)
    io.WriteString(w, `
<!DOCTYPE html>
<html>
  <head></head>
  <body>` +
   `<p>You are now logged in.</p>` +
   `<a href="/protectedpage">Test on this protected page</a>` +
`  </body>
</html>`)
}


//Login/logout
func login(res http.ResponseWriter, req *http.Request) {
    log.Print("saving logged in cookies")
    err := SetCookie(res, req, map[string]string{"loggedin": "true",})
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

//Authentication/page protection middleware
func authenticatePage(f http.HandlerFunc) http.HandlerFunc {
    return func(res http.ResponseWriter, req *http.Request) {
        log.Print("starting authentication")
        loggedin, err := GetCookie(req, "loggedin")
        if err != nil {
            log.Fatal("Failed to open cookie")
        }
        if loggedin == "false" || loggedin == "" { //return error/login page
            log.Print("authentication failed")
            log.Print(loggedin)
            failedLoginHandler(res, req)
        } else { //return page
            log.Print("authentication complete")
            f(res, req)
        }
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

func GetCookie(r *http.Request, key string) (string, error) {
    if cookie, err := r.Cookie(config.CookieName); err == nil {
        value := make(map[string]string)
        if err = store.Decode(config.CookieName, cookie.Value, &value); err == nil {
            return value[key], err
        } else {
            log.Print("failed to decode cookie")
            return "", err
        }
    } else {
        log.Print("failed to retrieve cookie")
        return "", err
    }
}

