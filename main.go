package main

import (
	"encoding/json"
	"flag"
	"github.com/go-resty/resty/v2"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"html/template"
	"net/http"
)

type toBePassed struct {
	Name string
	Keys Response
}

type Response []struct {
	ID  int    `json:"id"`
	Key string `json:"key"`
}

type Client struct {
	Logger    *logrus.Logger
	Templates *template.Template
	R         *resty.Client
}

func getKeysFromGitHub(client *resty.Client, user string) Response {
	resp, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Accept", "application/json").
		SetResult(Response{}).
		Get("https://api.github.com/users/" + user + "/keys")
	if err != nil {
		panic(err)
	}
	var f Response
	err = json.Unmarshal(resp.Body(), &f)
	return f
}

func (c *Client) handlerScript(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	user := vars["user"]
	response := getKeysFromGitHub(c.R, user)
	structToPass := toBePassed{Name: user, Keys: response}
	err := c.Templates.ExecuteTemplate(w, "script", structToPass)

	if err != nil {
		panic(err)
	}
}

func main() {

	templates := template.Must(template.New("").ParseGlob("*.template"))
	port := flag.String("listen", "0.0.0.0:8088", "address to listen on. e.g. 127.0.0.1:8080 or :8080")
	flag.Parse()

	logger := logrus.New()

	client := Client{Logger: logger, Templates: templates, R: resty.New()}

	r := mux.NewRouter()

	r.HandleFunc("/{user}", client.handlerScript)

	logger.Info("Listening on " + *port + "...")
	logger.Fatalln(http.ListenAndServe(*port, r))

}
