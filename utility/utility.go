package utility

func SplitByN(s string, n int) []string {
	var result []string
	runes := []rune(s)
	for i := 0; i < len(runes); i += n {
		end := min(i + n, len(runes))
		result = append(result, string(runes[i:end]))
	}
	return result
}
