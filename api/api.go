package api

import "time"

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
