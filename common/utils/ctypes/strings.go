package ctypes

import (
	"strconv"
	"strings"
)

func SubstringByMaxChars(s string, max int) string {
	words := strings.Fields(s)
	if len(words) == 0 {
		return ""
	}

	// Si la primera palabra es mayor al límite, retorna solo los primeros 'max' caracteres.
	if len(words[0]) > max {
		return words[0][:max]
	}

	// Inicializa con la primera palabra.
	result := words[0]
	currentLength := len(result)

	// Recorre las palabras restantes.
	for i := 1; i < len(words); i++ {
		word := words[i]
		// Se considera el espacio adicional antes de la siguiente palabra.
		if currentLength+1+len(word) <= max {
			result += " " + word
			currentLength += 1 + len(word)
		} else {
			break
		}
	}

	return result
}

func StringToFloat(s string) float64 {
	if strings.Contains(s, ",") {
		s = strings.ReplaceAll(s, ",", ".")
	}
	num, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0
	}
	return num
}

func MoneyToCents(money float64) int {
	return int(money * 100)
}

func StringToCents(s string) int {
	if strings.Contains(s, ",") {
		s = strings.ReplaceAll(s, ",", ".")
	}
	num, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0
	}
	return MoneyToCents(num)
}

func CentsToString(cents int) string {
	if cents == 0 {
		return "0.00"
	}
	return strconv.FormatFloat(float64(cents)/100, 'f', 2, 64)
}
