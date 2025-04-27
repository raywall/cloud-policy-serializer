package extractor

// Analisar o caminho em segmentos, lidando corretamente com os colchetes
func parsePath(path string) []string {
	var segments []string
	current := ""
	inBracket := false

	for _, char := range path {
		switch char {
		case '.':
			if !inBracket {
				if current != "" {
					segments = append(segments, current)
					current = ""
				}
			} else {
				current += string(char)
			}
		case '[':
			inBracket = true
			current += string(char)
		case ']':
			inBracket = false
			current += string(char)
		default:
			current += string(char)
		}
	}

	if current != "" {
		segments = append(segments, current)
	}

	return segments
}
