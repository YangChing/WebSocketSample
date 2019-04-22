package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/gorilla/websocket"
	"github.com/marcusolsson/tui-go"
)

type post struct {
	Username string `json:"username,omitempty"`
	Message  string `json:"message"`
	Time     string `json:"time"`
}

var posts = []post{post{Username: "info", Message: "use 'esc' to leave room ", Time: ""}}

var userName string

func main() {

	scanner := bufio.NewScanner(os.Stdin)
	fmt.Print("Enter your name: ")
	scanner.Scan()
	name := scanner.Text()
	userName = name
	fmt.Print("Entet ip:")
	scanner.Scan()
	ip := scanner.Text()
	if ip == "" {
		ip = "12345"
	}
	var addr = flag.String("addr", fmt.Sprintf(":%v", ip), "http service address")

	u := url.URL{Scheme: "ws", Host: *addr, Path: "/ws"}
	var dialer *websocket.Dialer

	h := http.Header{}
	h.Set("username", name)
	conn, _, err := dialer.Dial(u.String(), h)
	if err != nil {
		fmt.Println(err)
		return
	}

	ui, input, history := drawChatView()

	go senMessage(conn, input)
	go acceptMessage(conn, ui, history)
	// go creatView(ui)

	ui.SetKeybinding("Esc", func() { ui.Quit() })
	if err := ui.Run(); err != nil {
		log.Fatal(err)
	}

}

func senMessage(conn *websocket.Conn, input *tui.Entry) {

	input.OnSubmit(func(e *tui.Entry) {
		p := post{Username: userName, Time: time.Now().Format("2006-01-02 15:04:05"), Message: e.Text()}

		err := conn.WriteJSON(p)
		if err != nil {
			fmt.Println("json err:", err)
			return
		}

		input.SetText("")
	})
}

func acceptMessage(conn *websocket.Conn, ui tui.UI, history *tui.Box) {
	for {
		var p post
		err := conn.ReadJSON(&p)
		if err != nil {
			ui.Update(func() {
				history.Append(tui.NewHBox(
					tui.NewLabel(err.Error()),
					tui.NewSpacer(),
				))
			})
			return
		}

		ui.Update(func() {
			switch p.Message {
			case "entry room":
				history.Append(tui.NewHBox(
					tui.NewLabel(fmt.Sprintf("%v [ %v ] %v", p.Time, p.Username, p.Message)),
					tui.NewSpacer(),
				))
			case "leave room":
				history.Append(tui.NewHBox(
					tui.NewLabel(fmt.Sprintf("%v [ %v ] %v", p.Time, p.Username, p.Message)),
					tui.NewSpacer(),
				))
			default:
				history.Append(tui.NewHBox(
					tui.NewLabel(fmt.Sprintf("%v %v: %v", p.Time, p.Username, p.Message)),
					tui.NewSpacer(),
				))
			}
		})
	}
}

func drawChatView() (tui.UI, *tui.Entry, *tui.Box) {

	history := tui.NewVBox()

	for _, m := range posts {
		history.Append(tui.NewHBox(
			tui.NewLabel(m.Time),
			tui.NewPadder(1, 0, tui.NewLabel(fmt.Sprintf("<%s>", m.Username))),
			tui.NewLabel(m.Message),
			tui.NewSpacer(),
		))
	}

	historyScroll := tui.NewScrollArea(history)
	historyScroll.SetAutoscrollToBottom(true)

	historyBox := tui.NewVBox(historyScroll)
	historyBox.SetBorder(true)

	input := tui.NewEntry()
	input.SetFocused(true)
	input.SetSizePolicy(tui.Expanding, tui.Maximum)

	inputBox := tui.NewHBox(input)
	inputBox.SetBorder(true)
	inputBox.SetSizePolicy(tui.Expanding, tui.Maximum)

	chat := tui.NewVBox(historyBox, inputBox)
	chat.SetSizePolicy(tui.Expanding, tui.Expanding)
	chat.SetBorder(true)

	ui, err := tui.New(chat)
	if err != nil {
		log.Fatal(err)
	}

	return ui, input, history
}
