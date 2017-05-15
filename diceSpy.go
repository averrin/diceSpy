package main

import (
	"encoding/json"
	"fmt"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"io/ioutil"
	"net/http"
	"strings"
)

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
	OrigRoll   string
}

type RollWrapper struct {
	P string `json:"p"`
	D struct {
		Content  string `json:"content"`
		OrigRoll string `json:"origRoll"`
		Playerid string `json:"playerid"`
		Type     string `json:"type"`
		Who      string `json:"who"`
	} `json:"d"`
}

var players map[string]string

func main() {
	e := echo.New()
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{echo.GET, echo.PUT, echo.POST, echo.DELETE},
	}))
	e.File("/", "payload.js")

	e.POST("/players", func(c echo.Context) error {
		readPlayers(c.Request())
		fmt.Println(players)
		return c.String(http.StatusOK, "OK")
	})

	e.POST("/roll", func(c echo.Context) error {
		roll := readRoll(c.Request())
		fmt.Println(roll)
		results := roll.Rolls[0].Results
		skill := strings.TrimSpace(roll.Rolls[len(roll.Rolls)-1].Text)
		message := fmt.Sprintf("%v:", roll.Player)
		if skill != "" {
			message += fmt.Sprintf("\n%v", skill)
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
			mod := strings.TrimSpace(roll.Rolls[len(roll.Rolls)-2].Expr)
			if mod != "" {
				message += fmt.Sprintf(" %v", mod)
			}
		}
		message += fmt.Sprintf(" = %v", roll.Total)
		fmt.Println(
			ioutil.WriteFile("roll.txt",
				[]byte(message), 0644))
		return c.String(http.StatusOK, "OK")
	})
	e.Logger.Fatal(e.Start(":1323"))
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
	return &r
}

func readPlayers(req *http.Request) {
	decoder := json.NewDecoder(req.Body)
	decoder.Decode(&players)
}
