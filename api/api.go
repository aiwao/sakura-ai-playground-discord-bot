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
		return Message{}, err
	}

	var resPayload ResponsePayload
	if err := json.Unmarshal(b, &resPayload); err != nil {
		return Message{}, err
	}

	if len(resPayload.Content) > 0 {
		return resPayload.Content[len(resPayload.Content)-1], nil
	}

	return Message{}, errors.New("No messages was returned")
}

type AIModel int

const(
	GPT_OSS_120b AIModel = iota
	Qwen3_Coder_30B_A3B_Instruct
	Qwen3_Coder_480B_A35B_Instruct_FP8
	LLM_JP_3_1_8x13b_instruct4
	Phi_4_mini_instruct_cpu
	Phi_4_multimodal_instruct
	Qwen3_0_6B_cpu
	Qwen3_VL_30B_A3B_Instruct
)

func (a AIModel) Name() string {
	switch a {
	case GPT_OSS_120b:
		return "gpt-oss-120b"
	case Qwen3_Coder_30B_A3B_Instruct:
		return "Qwen3-Coder-30B-A3B-Instruct"
	case Qwen3_Coder_480B_A35B_Instruct_FP8:
		return "Qwen3-Coder-480B-A35B-Instruct-FP8"
	case LLM_JP_3_1_8x13b_instruct4:
		return "llm-jp-3.1-8x13b-instruct4"
	case Phi_4_mini_instruct_cpu:
		return "Phi-4-mini-instruct-cpu"
	case Phi_4_multimodal_instruct:
		return "Phi-4-multimodal-instruct"
	case Qwen3_0_6B_cpu:
		return "Qwen3-0-6B-cpu"
	case Qwen3_VL_30B_A3B_Instruct:
		return "Qwen3-VL-30B-A3B-Instruct"
	default:
		return ""
	}
}
