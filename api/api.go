package api

import (
	"errors"
	"net/http"
	"net/http/cookiejar"
	"regexp"
	"strconv"
	"time"

	instaddr "github.com/aiwao/instaddr_api"
	"github.com/aiwao/rik"
	"github.com/corpix/uarand"
)

const baseURL = "https://playground.aipf.sakura.ad.jp/api/"
const authURL = baseURL+"auth/"
const methodURL = authURL+"method/"
const loginURL = authURL+"login/"
const codeURL = loginURL+"code/"
const v1URL = baseURL+"v1/"
const userURL = v1URL+"user/"

type SakuraID struct {
	Email string `json:email`
	Password string `json:password`
	CreatedAt time.Time `json:created_at`
	InstaddrID string `json:instaddr_id`
	InstaddrPassword string `json:instaddr_password`
}

type SakuraSession struct {
	ID SakuraID
	CSRFToken string
	Jar *cookiejar.Jar
}

func (id SakuraID) NewSakuraSession() (*SakuraSession, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}
	client := &http.Client{
		Jar: jar,
	}
	acc, err := instaddr.LoginAccount(instaddr.Options{}, instaddr.AuthInfo{AccountID: id.InstaddrID, Password: id.InstaddrPassword})
	if err != nil {
		return nil, err
	}

	ua := uarand.GetRandom()
	_, err = rik.Post(methodURL).
		JSON(rik.NewJSON().Set("email", id.Email).Build()).
		Header("User-agent", ua).
		Header("X-Csrftoken", "undefined").
		DoClient(client)
	if err != nil {
		return nil, err
	}
	
	_, err = rik.Post(loginURL).
		JSON(rik.NewJSON().
			Set("email", id.Email).
			Set("password", id.Password).
			Build(),
		).
		Header("User-agent", ua).
		Header("X-Csrftoken", "undefined").
		DoClient(client)
	if err != nil {
		return nil, err
	}

	verifyCode := ""
	for range 20 {
		previews, err := acc.SearchMail(instaddr.SearchOptions{Query: id.Email})
		if err != nil {
			goto SLEEP
		}
		for _, p  := range previews {
			mail, err := acc.ViewMail(instaddr.Options{}, p)
			if err != nil {
				goto SLEEP
			}

			re := regexp.MustCompile(`\d+`)
			match := re.FindAllString(mail.Content, -1)
			if len(match) > 0 {
				_, err = strconv.Atoi(match[len(match)-1])
				if err != nil {
					goto SLEEP
				}
				verifyCode = match[len(match)-1]
				goto DONE
			}
		}
		SLEEP:
			time.Sleep(1*time.Second)
		DONE:
			if verifyCode != "" {
				break
			}
	}
	if verifyCode == "" {
		return nil, errors.New("failed to verify")
	}
	
	res, err := rik.Post(codeURL).
		JSON(rik.NewJSON().Set("code", verifyCode).Build()).
		Header("User-agent", ua).
		Header("X-Csrftoken", "undefined").
		DoClient(client)
	if err != nil {
		return nil, err
	}

	csrf := ""
	for _, c := range res.Cookies() {
		if c.Name == "csrftoken" {
			csrf = c.Value
			break
		}
	}
	if csrf == "" {
		return nil, errors.New("failed to verify")
	}

	return &SakuraSession{
		ID: id,
		CSRFToken: csrf,
		Jar: jar,
	}, nil
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
		return "Qwen3-Coder-480B-A35B-Instruct_FP8"
	case LLM_JP_3_1_8x13b_instruct4:
		return "llm-jp-3.1-8x13b-instruct4"
	case Phi_4_mini_instruct_cpu:
		return "Phi-4-mini-instruct-cpu"
	case Phi_4_multimodal_instruct:
		return "Phi-4-multimodal-instruct"
	case Qwen3_0_6B_cpu:
		return "Qwen3_0_6B_cpu"
	case Qwen3_VL_30B_A3B_Instruct:
		return "Qwen3-VL-30B-A3B-Instruct"
	default:
		return ""
	}
}
