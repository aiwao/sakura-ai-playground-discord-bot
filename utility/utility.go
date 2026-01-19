package utility

import (
	"log"
	"sakura_ai_bot/api"
	"sakura_ai_bot/environment"
)

func LoadSessionIDList() []api.SakuraID {
	sakuraIDList := []api.SakuraID{}
	rows, err := environment.DB.Query("SELECT email, password, instaddr_id, instaddr_password FROM accounts WHERE activate_at < now()")
	if err != nil {
		log.Fatalln(err)
	}
	defer rows.Close()

	for rows.Next() {
		var idScan api.SakuraID
		if err := rows.Scan(&idScan.Email, &idScan.Password, &idScan.InstaddrID, &idScan.InstaddrPassword); err != nil {
			log.Println(err)
			continue
		}
		sakuraIDList = append(sakuraIDList, idScan)
	}
	log.Printf("Sakura ID count: %d\n", len(sakuraIDList))
	return sakuraIDList
}

func SplitByN(s string, n int) []string {
	var result []string
	runes := []rune(s)
	for i := 0; i < len(runes); i += n {
		end := min(i + n, len(runes))
		result = append(result, string(runes[i:end]))
	}
	return result
}
