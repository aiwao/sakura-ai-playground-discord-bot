package api

import (
	"encoding/json"
	"errors"
	"net/http"
		
	"github.com/aiwao/rik"
	"github.com/corpix/uarand"
)

const baseURL = "https://playground.aipf.sakura.ad.jp"
const apiURL = baseURL+"/api/"
const authURL = apiURL+"auth/"
const sessionURL = authURL+"session"
const providersURL = authURL+"providers"
const csrfURL = authURL+"csrf"
const signinURL = authURL+"signin"
const sakuraSigninURL = signinURL+"/sakura"
const idpURL = "https://secure.sakura.ad.jp/serviceidp/api/v1/"
const idpAuthURL = idpURL+"auth/"
const methodURL = idpAuthURL+"method/"
const loginURL = idpAuthURL+"login/"
const codeURL = loginURL+"code/"
const userURL = idpURL+"user/"
const chatURL = apiURL+"chat/"

type Message struct {
	ID string `json:"id"`
	Role string `json:"role"`
	Content string `json:"content"`
}

type ChatPayload struct {
	Messages []Message `json:"messages"`
	Model string `json:"model"`
}

type ResponsePayload struct {
	Status string `json:"status"`
	Content []Message `json:"content"`
	Text string `json:"text"`
	TokenEnded bool `json:"token_ended"`
}

func (s *SakuraSession) Chat(payload ChatPayload) (Message, error) {
	client := &http.Client{
		Jar: s.Jar,
	}

	b, _, err := rik.Post(chatURL).
		JSON(payload).
		Header("User-Agent", uarand.GetRandom()).
		DoReadByteClient(client)
	if err != nil {
		s.InvalidRequestCount++
		return Message{}, err
	}

	var resPayload ResponsePayload
	if err := json.Unmarshal(b, &resPayload); err != nil {
		s.InvalidRequestCount++
		return Message{}, err
	}

	if len(resPayload.Content) > 0 {
		return resPayload.Content[len(resPayload.Content)-1], nil
	}

	s.InvalidRequestCount++
	return Message{}, errors.New("No messages was returned")
}

var AIModelList = []string{
	"gpt-oss-120b",
	"Qwen3-Coder-30B-A3B-Instruct",
	"Qwen3-Coder-480B-A35B-Instruct-FP8",
	"llm-jp-3.1-8x13b-instruct4",
	"preview/Phi-4-mini-instruct-cpu",
	"preview/Phi-4-multimodal-instruct",
	"preview/Qwen3-0.6B-cpu",
	"preview/Qwen3-VL-30B-A3B-Instruct",
}
