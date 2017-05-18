package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/golang/freetype"
	"github.com/jinzhu/configor"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/lucasb-eyer/go-colorful"
	"golang.org/x/image/font"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
)

const avatarRoot string = "https://app.roll20.net"

var Config = struct {
	HistoryCount int `default:"1"`
	Image        struct {
		FontSize float64 `default:"16"`
		Dpi      float64 `default:"144"`
		FontFile string  `default:"Monofonto"`
		Color    string  `default:"1fd6ef"`
		Width    int     `default:"144"`
		Height   int     `default:"144"`
	}
}{}

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

var rolls []*Roll

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
	configor.Load(&Config, "config.yml")
	fmt.Println(Config)
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
		if len(rolls) == Config.HistoryCount {
			rolls = rolls[1:]
		}
		rolls = append(rolls, roll)
		message := ""
		for _, r := range rolls {
			message += renderRoll(r) + "\n\n"
		}

		fmt.Println(
			ioutil.WriteFile("roll.txt",
				[]byte(message), 0644))
		drawText(message)
		return c.String(http.StatusOK, "OK")
	})
	e.Logger.Fatal(e.Start(":1323"))
}

func renderRoll(roll *Roll) string {
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
	return &r
}

func readPlayers(req *http.Request) {
	decoder := json.NewDecoder(req.Body)
	decoder.Decode(&players)
}

func drawText(text string) {

	// Read the font data.
	fontBytes, err := ioutil.ReadFile(Config.Image.FontFile + ".ttf")
	if err != nil {
		log.Println(err)
		return
	}
	f, err := freetype.ParseFont(fontBytes)
	if err != nil {
		log.Println(err)
		return
	}

	// Initialize the context. 1fd6ef
	cl, err := colorful.Hex(Config.Image.Color)
	fmt.Println(cl.R, cl.G, cl.B)
	fg := image.NewUniform(color.RGBA{uint8(cl.R * 255), uint8(cl.G * 255), uint8(cl.B * 255), 0xff})
	bg := image.Transparent
	rgba := image.NewRGBA(image.Rect(0, 0, Config.Image.Width, Config.Image.Height))
	draw.Draw(rgba, rgba.Bounds(), bg, image.ZP, draw.Src)
	c := freetype.NewContext()
	c.SetDPI(Config.Image.Dpi)
	c.SetFont(f)
	c.SetFontSize(Config.Image.FontSize)
	c.SetClip(rgba.Bounds())
	c.SetDst(rgba)
	c.SetSrc(fg)
	c.SetHinting(font.HintingFull)

	// Draw the text.
	pt := freetype.Pt(10, 10+int(c.PointToFixed(Config.Image.FontSize)>>6))
	for _, s := range strings.Split(text, "\n") {
		_, err = c.DrawString(s, pt)
		if err != nil {
			log.Println(err)
			return
		}
		pt.Y += c.PointToFixed(Config.Image.FontSize * 1)
	}

	// Save that RGBA image to disk.
	outFile, err := os.Create("roll.png")
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	defer outFile.Close()
	b := bufio.NewWriter(outFile)
	err = png.Encode(b, rgba)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	err = b.Flush()
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	fmt.Println("Wrote out.png OK.")
}
