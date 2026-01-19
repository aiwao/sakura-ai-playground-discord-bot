package api

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"net/http/cookiejar"
	"regexp"
	"sakura_ai_bot/environment"
	"strconv"
	"time"

	instaddr "github.com/aiwao/instaddr_api"
	"github.com/aiwao/rik"
	"github.com/corpix/uarand"
)

type SakuraID struct {
	Email string `json:"email"`
	Password string `json:"password"`
	InstaddrID string `json:"instaddr_id"`
	InstaddrPassword string `json:"instaddr_password"`
}

type SakuraSession struct {
	ID SakuraID
	Jar *cookiejar.Jar
	InvalidRequestCount int
	Active bool
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

	b, res, err := rik.Get(csrfURL).
		Header("User-Agent", ua).
		Header("Content-Type", rik.ContentTypeJSON).
		DoReadByteClient(client)
	if err != nil && err != http.ErrUseLastResponse {
		return nil, err
	}
	log.Printf("Get CSRF token: %d\n", res.StatusCode)

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
	log.Printf("Start sign into Sakura ID: %d\n", res.StatusCode)

	var sakuraRes struct {
		URL string `json:"url"`
	}
	if err := json.Unmarshal(b, &sakuraRes); err != nil {
		return nil, err
	}

	res, err = rik.Get(sakuraRes.URL).
		Header("User-Agent", ua).
		DoClient(client)
	if err != nil && err != http.ErrUseLastResponse {
		return nil, err
	}
	log.Printf("Start OAuth: %d\n", res.StatusCode)

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
	log.Printf("Start OAuth login: %d\n", res.StatusCode)

	res, err = rik.Post(loginURL).
		JSON(rik.NewJSON().
			Set("email", id.Email).
			Set("password", id.Password).
			Build(),
		).
		Header("User-Agent", ua).
		Header("X-Csrftoken", "undefined").
		DoClient(client)
	if err != nil && err != http.ErrUseLastResponse {
		return nil, err
	}
	log.Printf("Login: %d\n", res.StatusCode)

	time.Sleep(5*time.Second)
	verifyCode := ""
	for range 20 {
		previews, err := acc.SearchMail(instaddr.SearchOptions{Query: id.Email})
		if err != nil {
			time.Sleep(time.Duration(environment.CheckMailDelay)*time.Millisecond)
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
		time.Sleep(time.Duration(environment.CheckMailDelay)*time.Millisecond)
	}
	if verifyCode == "" {
		return nil, errors.New("failed to verify")
	}
	log.Println("Received verification code: "+verifyCode)

	res, err = rik.Post(codeURL).
		JSON(rik.NewJSON().Set("code", verifyCode).Build()).
		Header("User-Agent", ua).
		Header("X-Csrftoken", "undefined").
		DoClient(client)
	if err != nil && err != http.ErrUseLastResponse {
		return nil, err
	}
	log.Printf("Code verification: %d\n", res.StatusCode)	

	res, err = rik.Get(sakuraRes.URL).
		Header("User-Agent", ua).
		DoClient(client)
	if err != nil && err != http.ErrUseLastResponse {
		return nil, err
	}
	log.Printf("OAuth: %d\n", res.StatusCode)
	
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
	log.Printf("OAuth login: %d\n", res.StatusCode)

	return &SakuraSession{
		ID: id,
		Jar: jar,
		InvalidRequestCount: 0,
		Active: true,
	}, nil
}
