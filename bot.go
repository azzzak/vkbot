package vkbot

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// BotAPI contains information for using API.
type BotAPI struct {
	Token  string `json:"token"`
	Buffer int    `json:"buffer"`

	GroupID      int    `json:"-"`
	Secret       string `json:"-"`
	Confirmation string `json:"-"`

	Client *http.Client `json:"-"`
}

// Packet is type for incoming message.
type Packet struct {
	Type    string  `json:"type"`
	Payload Payload `json:"object"`
	GroupID int     `json:"group_id"`
	Secret  string  `json:"secret"`
}

// Payload is type for payload of incoming message.
type Payload struct {
	ID        int    `json:"id"`
	Date      int64  `json:"date"`
	Out       int    `json:"out"`
	UserID    int    `json:"user_id"`
	ReadState int    `json:"read_state"`
	Title     string `json:"title"`
	Body      string `json:"body"`
	JoinType  int    `json:"join_type"`
	Self      int    `json:"self"`
}

// Response is type for response in Send function.
type Response struct {
	Error    Error `json:"error"`
	Response int   `json:"response"`
}

// Error is type for possibly error.
type Error struct {
	ErrorCode     int           `json:"error_code"`
	ErrorMsg      string        `json:"error_msg"`
	RequestParams []interface{} `json:"request_params"`
}

const (
	endpoint   = "https://api.vk.com/method"
	sendMethod = "messages.send"
)

// Constant values for type of message.
const (
	IncomingMessage  = "message_new"
	OutcomingMessage = "message_reply"
	JoinGroup        = "group_join"
	LeaveGroup       = "group_leave"
)

var currentTokenNum int

func init() {
	rand.Seed(time.Now().UnixNano())
}

func rotateTokens(tokens []string) string {
	defer func() {
		currentTokenNum++
		if currentTokenNum >= len(tokens) {
			currentTokenNum = 0
		}
	}()
	return tokens[currentTokenNum]
}

func randomID() string {
	return strconv.Itoa(rand.Intn(10000))
}

// NewBot creates a new BotAPI instance.
func NewBot(token string, group int) (*BotAPI, error) {
	httpClient := &http.Client{
		Timeout: time.Second * 30,
	}

	bot := &BotAPI{
		Token:   token,
		GroupID: group,
		Client:  httpClient,
		Buffer:  20,
	}

	return bot, nil
}

// ListenForWebhook registers a http handler for a webhook.
func (bot *BotAPI) ListenForWebhook(pattern string) <-chan Packet {
	ch := make(chan Packet, bot.Buffer)
	http.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		decoder := json.NewDecoder(r.Body)
		var p Packet
		err := decoder.Decode(&p)
		if err != nil {
			fmt.Println(err)
		}

		if p.Secret != bot.Secret {
			return
		}

		if p.Type == "confirmation" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(bot.Confirmation))
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
		ch <- p
	})
	return ch
}

// Send will send message to user through VK.
func (bot *BotAPI) Send(id int, text ...string) (Response, error) {
	message := strings.Join(text, "<br>")

	params := url.Values{}
	params.Add("user_id", fmt.Sprintf("%d", id))
	params.Add("random_id", randomID())
	params.Add("peer_id", fmt.Sprintf("-%d", bot.GroupID))
	params.Add("message", message)
	params.Add("access_token", bot.Token)
	params.Add("v", "5.67")

	url := fmt.Sprintf("%s/%s", endpoint, sendMethod)
	r, err := bot.Client.PostForm(url, params)
	if err != nil {
		fmt.Println(err)
	}
	defer r.Body.Close()

	var rp Response

	if r.StatusCode != http.StatusOK {
		return rp, errors.New(http.StatusText(r.StatusCode))
	}

	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&rp)
	if err != nil {
		fmt.Println(err)
	}

	if err := rp.Error.ErrorMsg; err != "" {
		return rp, errors.New(err)
	}

	return rp, nil
}
