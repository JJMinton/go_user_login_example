package main


import ("net/http"
        "io"
        "io/ioutil"
        "log"
        "os"

        "encoding/json"

        "golang.org/x/oauth2"
        "golang.org/x/oauth2/google"
        )

//Global vars
type Config struct {
    RootURL  string `json:"rootURL"`
    Port string `json:"port"`
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
    file, err := ioutil.ReadFile("./config.json")
    if err != nil {
        log.Printf("File error: %v\n", err)
        os.Exit(1)
    }
    json.Unmarshal(file, &config)

    //Google oauth config
    file, err = ioutil.ReadFile("./google_creds.json")
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
    http.ListenAndServe(config.Port, nil)
}


//Handlers
func rootHandler(res http.ResponseWriter, req *http.Request) {
    io.WriteString(res, `
<!DOCTYPE html>
<html>
  <head></head>
  <body>
    <a href="/auth/google/login">LOGIN WITH Google</a>
  </body>
</html>`)
}

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

    w.Write([]byte("token is " + tok.AccessToken + "\n"))

    response, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + tok.AccessToken)
    defer response.Body.Close()
    contents, err := ioutil.ReadAll(response.Body)
    if err != nil {
      log.Fatal("failed to read google respones body")
    }

    w.Write([]byte(contents))
}
