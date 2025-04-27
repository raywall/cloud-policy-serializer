package extractor

import (
	"regexp"
	"strconv"
)

// parsePath analisa o caminho em segmentos, lidando corretamente com os colchetes
func parsePath(path string) []pathSegment {
	segments := []pathSegment{}

	// Regex para detectar segmentos de caminho
	propPattern := regexp.MustCompile(`^\.?([^.\[\]]+)`)
	arrayPattern := regexp.MustCompile(`^\[(\d+)\]`)

	remaining := path

	for len(remaining) > 0 {
		// Tenta corresponder a uma propriedade
		if match := propPattern.FindStringSubmatch(remaining); len(match) > 0 {
			segments = append(segments, propertySegment{name: match[1]})
			remaining = remaining[len(match[0]):]
			continue
		}

		// Tenta corresponder a um índice de array
		if match := arrayPattern.FindStringSubmatch(remaining); len(match) > 0 {
			index, _ := strconv.Atoi(match[1])
			segments = append(segments, arrayIndexSegment{index: index})
			remaining = remaining[len(match[0]):]
			continue
		}

		// Se chegarmos aqui, o formato é inválido
		break
	}

	return segments
}

// func parsePath(path string) []string {
// 	var segments []string
// 	current := ""
// 	inBracket := false

// 	for _, char := range path {
// 		switch char {
// 		case '.':
// 			if !inBracket {
// 				if current != "" {
// 					segments = append(segments, current)
// 					current = ""
// 				}
// 			} else {
// 				current += string(char)
// 			}
// 		case '[':
// 			inBracket = true
// 			current += string(char)
// 		case ']':
// 			inBracket = false
// 			current += string(char)
// 		default:
// 			current += string(char)
// 		}
// 	}

// 	if current != "" {
// 		segments = append(segments, current)
// 	}

// 	return segments
// }
