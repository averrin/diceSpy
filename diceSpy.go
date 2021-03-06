package main

import (
	"encoding/json"
	"fmt"
	"github.com/jinzhu/configor"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"io"
	"io/ioutil"
	// "log"
	"golang.org/x/net/websocket"
	"net/http"
	"strings"
	"text/template"
)

const avatarRoot string = "https://app.roll20.net"

type ConfigStruct struct {
	HistoryCount int `default:"1"`
}

var Config = ConfigStruct{}

type Roll struct {
	Type  string `json:"type"`
	Rolls []struct {
		Type string `json:"type"`
		Dice int    `json:"dice,omitempty"`
		// Fate bool   `json:"fate,omitempty"`
		Mods struct {
		} `json:"mods,omitempty"`
		Sides   int `json:"sides,omitempty"`
		Results []struct {
			V int `json:"v"`
		} `json:"results,omitempty"`
		Expr string `json:"expr,omitempty"`
		Text string `json:"text,omitempty"`
	} `json:"rolls"`
	ResultType string `json:"resultType"`
	Total      int    `json:"total"`
	Player     string
	Avatar     string
	OrigRoll   string
	Message    string
	Skill      string
	Mod        string
	Results    []struct {
		V int `json:"v"`
	}
}

var rolls []*Roll

type RollWrapper struct {
	P string `json:"p"`
	D struct {
		Content  string `json:"content"`
		Avatar   string `json:"avatar"`
		OrigRoll string `json:"origRoll"`
		Playerid string `json:"playerid"`
		Type     string `json:"type"`
		Who      string `json:"who"`
	} `json:"d"`
}

var players map[string]string

type Template struct {
	templates *template.Template
}

func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

func result(c echo.Context) error {
	return c.Render(http.StatusOK, c.Param("name"), struct {
		Rolls  []*Roll
		Config ConfigStruct
	}{rolls, Config})
}

var socket *websocket.Conn

func wsHandler(c echo.Context) error {
	websocket.Handler(func(ws *websocket.Conn) {
		socket = ws
		defer ws.Close()
		for {
			websocket.Message.Receive(ws, nil)
		}
	}).ServeHTTP(c.Response(), c.Request())
	return nil
}

func main() {
	configor.Load(&Config, "config.yml")
	t := &Template{
		templates: template.Must(template.ParseGlob("templates/*.html")),
	}
	e := echo.New()
	e.Renderer = t
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{echo.GET, echo.PUT, echo.POST, echo.DELETE},
	}))
	e.File("/script", "payload.js")
	e.GET("/display/:name", result)
	e.GET("/ws", wsHandler)
	e.Static("/templates", "templates")

	e.POST("/players", func(c echo.Context) error {
		readPlayers(c.Request())
		fmt.Println(players)
		return c.String(http.StatusOK, "OK")
	})

	e.POST("/roll", func(c echo.Context) error {
		configor.Load(&Config, "config.yml")
		roll := readRoll(c.Request())
		fmt.Println(roll)
		for len(rolls) >= Config.HistoryCount {
			rolls = rolls[1:]
		}
		rolls = append(rolls, roll)
		message := ""
		for _, r := range rolls {
			r.Message = renderRoll(r)
			message += r.Message + "\n\n"
		}

		ioutil.WriteFile("roll.txt",
			[]byte(message), 0644)

		if socket != nil {
			websocket.Message.Send(socket, "Hello, Client!")
		}
		return c.String(http.StatusOK, "OK")
	})
	fmt.Println("")
	fmt.Println("-------")
	fmt.Println("")
	fmt.Println("Exec `$.getScript('http://127.0.0.1:1323/script');` in roll20.net WebInspector console")
	fmt.Println("Use `http://127.0.0.1:1323/display/basic` as OBS BrowserSource")
	fmt.Println("")
	fmt.Println("-------")
	fmt.Println("")
	e.Logger.Fatal(e.Start(":1323"))
}

func renderRoll(roll *Roll) string {
	results := roll.Rolls[0].Results
	roll.Results = results
	roll.Skill = strings.TrimSpace(roll.Rolls[len(roll.Rolls)-1].Text)
	message := fmt.Sprintf("%v:", roll.Player)
	if roll.Skill != "" {
		message += fmt.Sprintf("\n%v", roll.Skill)
	}
	message += "\n("
	for i, d := range results {
		if i < len(results)-1 {
			message += fmt.Sprintf("%v, ", d.V)
		} else {
			message += fmt.Sprintf("%v", d.V)
		}
	}
	message += ")"

	if len(roll.Rolls) >= 3 {
		roll.Mod = strings.TrimSpace(roll.Rolls[len(roll.Rolls)-2].Expr)
		if roll.Mod != "" {
			message += fmt.Sprintf(" %v", roll.Mod)
		}
	}
	message += fmt.Sprintf(" = %v", roll.Total)
	return message
}

func readRoll(req *http.Request) *Roll {
	decoder := json.NewDecoder(req.Body)
	var rw RollWrapper
	err := decoder.Decode(&rw)
	if err != nil {
		panic(err)
	}
	defer req.Body.Close()
	var r Roll
	err = json.Unmarshal([]byte(rw.D.Content), &r)
	r.Player = players[rw.D.Playerid]
	r.OrigRoll = rw.D.OrigRoll
	r.Avatar = fmt.Sprintf("%v/users/avatar/%v/200", avatarRoot, strings.Split(rw.D.Avatar, "/")[3])

	return &r
}

func readPlayers(req *http.Request) {
	decoder := json.NewDecoder(req.Body)
	decoder.Decode(&players)
}
