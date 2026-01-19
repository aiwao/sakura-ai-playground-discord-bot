package sessionmanager

import (
	"errors"
	"log"
	"sakura_ai_bot/api"
	"sakura_ai_bot/environment"
	"sakura_ai_bot/utility"
	"time"
)

type RequestMethod int

const (
	Get RequestMethod = iota
	Deactivate
	Reload
)

type RequestBody struct {
	Method RequestMethod
	AmountGet int
	EmailDeactivate string
	IDListReload []api.SakuraID
}

type ResponseStatus int

const (
	Error ResponseStatus = iota
	Success
)

type ResponseBody struct {
	Error error
	SessionGet []*api.SakuraSession
	Count int
}

type ChannelMessage struct {
	req RequestBody
	resCh chan ResponseBody
}

var ch = make(chan ChannelMessage)

func Request(req RequestBody) (ResponseBody, error) {
	resCh := make(chan ResponseBody)
	ch <- ChannelMessage{req: req, resCh: resCh}
	select {
		case res := <-resCh:
			if res.Error != nil {
				return res, res.Error
			}
			return res, nil
		case <-time.After(2*time.Second):
			return ResponseBody{}, errors.New("Timed out")
	}
}

var sessionList = []*api.SakuraSession{}
var sakuraIDList = []api.SakuraID{}

func StartServer(idList []api.SakuraID) {
	sakuraIDList = idList
	go func() {
		sessionCnt := 0
		for {
			select {
				case msg := <-ch:
					req := msg.req
					resCh := msg.resCh
					switch req.Method {
						case Get:
							resSessionList := []*api.SakuraSession{}
							addedCnt := 0
							for _, s := range sessionList {
								if s.Active {
									resSessionList = append(resSessionList, s)
									addedCnt++
									if addedCnt >= req.AmountGet {
										break
									}
								}
							}
							resCh <- ResponseBody{Error: nil, SessionGet: resSessionList}
						case Deactivate:
							responsed := false
							for _, s := range sessionList {
								if s.ID.Email == req.EmailDeactivate {
									responsed = true

									_, err := environment.DB.Exec(
										"UPDATE accounts SET activate_at = now() + INTERVAL '24 hours' WHERE email = $1",
										req.EmailDeactivate,
									)
									if err != nil {
										resCh <- ResponseBody{Error: err}
										break
									}

									s.Active = false
									
									resCh <- ResponseBody{Error: nil}
									break
								}
							}
							sakuraIDList = utility.LoadSessionIDList()
							sessionCnt--
							if !responsed {
								resCh <- ResponseBody{Error: errors.New("no account was found")}
							}
						case Reload:
							sessionList = []*api.SakuraSession{}
							sessionCnt = 0
							sakuraIDList = req.IDListReload
							resCh <- ResponseBody{Error: nil}
					}
				default:
					if sessionCnt < environment.MaxSessions && sessionCnt < len(sakuraIDList) {
						sessionCnt++
						id := sakuraIDList[sessionCnt]
						session, err := id.NewSakuraSession()
						if err != nil {
							log.Println(err)
							continue
						}
						sessionList = append(sessionList, session)
					}
			}
		}
	}()	
}
