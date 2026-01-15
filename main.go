package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/bwmarrin/discordgo"
	_ "github.com/lib/pq"
	"golang.org/x/text/cases"
)

type SakuraID struct {
	Email string
	Password string
	CreatedAt time.Time
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

const baseURL = "https://playground.aipf.sakura.ad.jp/api"
const chatURL = baseURL+"/chat"

func main() {
	dbHost := os.Getenv("ID_DB_HOST")	
	dbPort := os.Getenv("ID_DB_PORT")	
	dbUser := os.Getenv("ID_DB_USER")	
	dbPass := os.Getenv("ID_DB_PASS")	
	dbName := os.Getenv("ID_DB_NAME")
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", dbHost, dbPort, dbUser, dbPass, dbName)
	log.Println("Database: " + dsn)
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalln(err)
	}
	defer db.Close()

	amountSakuraIDToUseEnv := os.Getenv("AMOUNT_SAKURA_ID_TO_USE")
	amountSakuraIDToUse, err := strconv.Atoi(amountSakuraIDToUseEnv)
	if err != nil {
		log.Println(err)
		amountSakuraIDToUse = 10
	}
	log.Printf("Use %d Sakura IDs\n", amountSakuraIDToUse)

	sakuraIDCount := 0
	sakuraIDList := []SakuraID{}
	for rows.Next() {
		var sakuraID SakuraID
		if err := rows.Scan(&sakuraID.Email, &sakuraID.Password, &sakuraID.CreatedAt); err != nil {
			log.Println(err)
			continue
		}
		sakuraIDList = append(sakuraIDList, account)
		sakuraIDCount++
		if sakuraIDCount >= amountSakuraIDToUse {
			break
		}
	}
	if len(sakuraIDList) == 0 {
		log.Fatalln("No Sakura ID is found.")
	}
	log.Printf("Sakura ID count: %d\n", len(sakuraIDList))

	s, err := discordgo.New("Bot "+os.Getenv("BOT_TOKEN"))
	if err != nil {
		log.Fatalln(err)
	}

	s.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		log.Printf("Logged in as: %v#%v", s.State.User.Username, s.State.User.Discriminator)
	})
	s.ApplicationCommandCreate(
		s.State.Application.ID,
		"",
		&discordgo.ApplicationCommand{
			Name:        "ask",
			Description: "Ask the AI",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "message",
					Description: "Message",
					Required:    true,
				},
				{
					Type: discordgo.ApplicationCommandOptionInteger,
					Name: "model",
					Description: "AI model to ask",
					Required: true,
					Choices: []*discordgo.ApplicationCommandOptionChoice{
						{
							Name:  GPT_OSS_120b.Name(),
							Value: GPT_OSS_120b,
						},
						{
							Name:  Qwen3_Coder_30B_A3B_Instruct.Name(),
							Value: Qwen3_Coder_30B_A3B_Instruct,
						},
						{
							Name:  Qwen3_Coder_480B_A35B_Instruct_FP8.Name(),
							Value: Qwen3_Coder_480B_A35B_Instruct_FP8,
						},
						{
							Name: LLM_JP_3_1_8x13b_instruct4.Name(), 
							Value: LLM_JP_3_1_8x13b_instruct4,
						},
						{
							Name:  Phi_4_mini_instruct_cpu.Name(),
							Value: Phi_4_mini_instruct_cpu,
						},
						{
							Name: Phi_4_multimodal_instruct.Name(),
							Value: Phi_4_multimodal_instruct,
						},
						{
							Name:  Qwen3_0_6B_cpu.Name(),
							Value: Qwen3_0_6B_cpu,
						},
						{
							Name:  Qwen3_VL_30B_A3B_Instruct.Name(),
							Value: Qwen3_VL_30B_A3B_Instruct,
						},
					},
				},
			},
		},	
	)
	
	s.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if i.ApplicationCommandData().Name == "ask" {
			options := i.ApplicationCommandData().Options
			optionMap := make(map[string]*discordgo.ApplicationCommandInteractionDataOption, len(options))
			for _, o  := range options {
				optionMap[o.Name] = o
			}

			msg := ""
			if o, ok := optionMap["message"]; ok {
				msg = o.StringValue()
			}
			model := GPT_OSS_120b
			if o, ok := optionMap["model"]; ok {
				model = AIModel(o.IntValue())
			}
			if msg == "" {
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{Content: "Message is must to be not empty"},
				})
				return
			}
		}
	})
}
