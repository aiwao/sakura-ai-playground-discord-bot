package api

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"net/http/cookiejar"
	"regexp"
	"strconv"
	"time"

	instaddr "github.com/aiwao/instaddr_api"
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

type SakuraID struct {
	Email string `json:"email"`
	Password string `json:"password"`
	CreatedAt time.Time `json:"created_at"`
	InstaddrID string `json:"instaddr_id"`
	InstaddrPassword string `json:"instaddr_password"`
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
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
        	return http.ErrUseLastResponse
    	},
	}
	acc, err := instaddr.LoginAccount(instaddr.Options{}, instaddr.AuthInfo{AccountID: id.InstaddrID, Password: id.InstaddrPassword})
	if err != nil {
		return nil, err
	}

	ua := uarand.GetRandom()

	s, res, err := rik.Get(sessionURL).
		Header("User-Agent", ua).
		Header("Content-Type", rik.ContentTypeJSON).
		DoReadStringClient(client)
	if err != nil && err != http.ErrUseLastResponse {
		return nil, err
	}
	log.Println("ses")
	log.Println(res.StatusCode)
	log.Println(s)
	log.Println(res.Cookies())	

	s, res, err = rik.Get(providersURL).
		Header("User-Agent", ua).
		Header("Content-Type", rik.ContentTypeJSON).
		DoReadStringClient(client)
	if err != nil && err != http.ErrUseLastResponse {
		return nil, err
	}
	log.Println("pro")
	log.Println(res.StatusCode)
	log.Println(s)

	b, res, err := rik.Get(csrfURL).
		Header("User-Agent", ua).
		Header("Content-Type", rik.ContentTypeJSON).
		DoReadByteClient(client)
	if err != nil && err != http.ErrUseLastResponse {
		return nil, err
	}
	log.Println("csrf")
	log.Println(res.StatusCode)
	log.Println(string(b))

	var csrfRes struct {
		CSRFToken string `json:"csrfToken"`
	}
	if err := json.Unmarshal(b, &csrfRes); err != nil {
		return nil, err
	}
	
	b, res, err = rik.Post(sakuraSigninURL).
		Form("csrfToken", csrfRes.CSRFToken).
		Form("callbackUrl", baseURL).
		Form("json", "true").	
		Header("User-Agent", ua).
		DoReadByteClient(client)
	if err != nil && err != http.ErrUseLastResponse {
		return nil, err
	}
	log.Println("sakura")
	log.Println(string(b))
	log.Println(res.Cookies())
	log.Println(res.StatusCode)
	log.Println(res.Header)

	var sakuraRes struct {
		URL string `json:"url"`
	}
	if err := json.Unmarshal(b, &sakuraRes); err != nil {
		return nil, err
	}

	s, res, err = rik.Get(sakuraRes.URL).
		Header("User-Agent", ua).
		DoReadStringClient(client)
	log.Println("AA")
	log.Println(res.Header)
	log.Println(s)
	log.Println(res.Cookies())
	log.Println(res.StatusCode)

	loc, err := res.Location()
	if err != nil {
		return nil, err
	}
	res, err = rik.Get(loc.String()).
		Header("User-Agent", ua).
		DoClient(client)
	if err != nil && err != http.ErrUseLastResponse {
		return nil, err
	}
	log.Println("Loggg auth")
	log.Println(res.StatusCode)
	log.Println(res.Header)
	
	s, res, err = rik.Post(methodURL).
		JSON(rik.NewJSON().Set("email", id.Email).Build()).
		Header("User-Agent", ua).
		Header("X-Csrftoken", "undefined").
		DoReadStringClient(client)
	if err != nil && err != http.ErrUseLastResponse {
		return nil, err
	}
	log.Println("method")
	log.Println(s)
	log.Println(res.StatusCode)
	
	s, res, err = rik.Post(loginURL).
		JSON(rik.NewJSON().
			Set("email", id.Email).
			Set("password", id.Password).
			Build(),
		).
		Header("User-Agent", ua).
		Header("X-Csrftoken", "undefined").
		DoReadStringClient(client)
	if err != nil && err != http.ErrUseLastResponse {
		return nil, err
	}
	log.Println("login")
	log.Println(s)
	log.Println(res.StatusCode)

	time.Sleep(5*time.Second)
	verifyCode := ""
	for range 20 {
		previews, err := acc.SearchMail(instaddr.SearchOptions{Query: id.Email})
		if err != nil {
			time.Sleep(1*time.Second)
			continue
		}
		
		maxID := 0
		for _, p  := range previews {		
			mailID, err := strconv.Atoi(p.MailID)
			if err != nil {
				continue
			}
			if mailID > maxID {
				maxID = mailID
				mail, err := acc.ViewMail(instaddr.Options{}, p)
				if err != nil {
					continue
				}

				re := regexp.MustCompile(`\d+`)
				match := re.FindAllString(mail.Content, -1)
				if len(match) > 0 {
					_, err = strconv.Atoi(match[len(match)-1])
					if err != nil {
						continue
					}
					verifyCode = match[len(match)-1]
				}
			}
		}
		
		if verifyCode != "" {
			break
		}
		time.Sleep(1*time.Second)
	}
	if verifyCode == "" {
		return nil, errors.New("failed to verify")
	}
	log.Println(verifyCode)

	s, res, err = rik.Post(codeURL).
		JSON(rik.NewJSON().Set("code", verifyCode).Build()).
		Header("User-Agent", ua).
		Header("X-Csrftoken", "undefined").
		DoReadStringClient(client)
	if err != nil && err != http.ErrUseLastResponse {
		return nil, err
	}
	log.Println(s)

	csrf := ""
	log.Println(res.Cookies())
	for _, c := range res.Cookies() {
		if c.Name == "csrftoken" {
			csrf = c.Value
			break
		}
	}
	if csrf == "" {
		return nil, errors.New("failed to get csrf token")
	}

	s, res, err = rik.Get(userURL).
		Header("User-Agent", ua).
		Header("X-Csrftoken", csrf).
		DoReadStringClient(client)
	if err != nil && err != http.ErrUseLastResponse {
		return nil, err
	}
	log.Println("user")
	log.Println(res.StatusCode)
	log.Println(s)
		
	res, err = rik.Get(sakuraRes.URL).
		Header("User-Agent", ua).
		DoClient(client)
	if err != nil && err != http.ErrUseLastResponse {
		return nil, err
	}
	log.Println("authhh")
	log.Println(res.StatusCode)
	log.Println(res.Header)
	
	locLoginAuth, err := res.Location()
	if err != nil {
		return nil, err
	}
	res, err = rik.Get(locLoginAuth.String()).
		Header("User-Agent", ua).
		DoClient(client)
	if err != nil && err != http.ErrUseLastResponse {
		return nil, err
	}
	log.Println("login authh")
	log.Println(res.StatusCode)
	log.Println(res.Header)

	locErrPage, err := res.Location()
	if err == nil {
		res, err := rik.Get(locErrPage.String()).
			Header("User-Agent", ua).
			DoClient(client)
		if err != nil && err != http.ErrUseLastResponse {
			return nil, err
		}
		log.Println("err page")
		log.Println(res.StatusCode)
		log.Println(res.Header)
	}

	return &SakuraSession{
		ID: id,
		CSRFToken: csrf,
		Jar: jar,
	}, nil
}

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

	b, res, err := rik.Post(chatURL).
		JSON(payload).
		Header("User-Agent", uarand.GetRandom()).
		DoReadByteClient(client)
	if err != nil {
		return Message{}, err
	}

	log.Println("CHATTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTT")
	log.Println(res.StatusCode)
	log.Println(string(b))
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
