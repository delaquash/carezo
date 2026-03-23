package utils

import (
	"strings"
	"unicode"
)

// SplitFullName splits a full name into first name and last name
// Examples:
//   "John Doe" -> ("John", "Doe")
//   "Mary Jane Smith" -> ("Mary", "Jane Smith")

func SplitFullName(fullName string) (string, string){
	// trim whitespace
	fullName= strings.TrimSpace(fullName)

	if fullName == "" {
		return "", ""
	}

	// split by whitespace 
	names:= strings.Fields(fullName)

	if len(names) == 0 {
		return "", ""
	}

	// first word is first name, rest = lastname
	firstName := names[0]
	lastName := strings.Join(names[1:], " ")

	return firstName, lastName
}

// Capitalize first lette of each name
// Example: "john doe" -> "John Doe"

func CapitalizeName(name string) string {
	words:= strings.Fields(name)
	for i, word:= range words {
		if len(word) > 0 {
			runes := []rune(word)
			runes[0] = unicode.ToUpper(runes[0])
			words[i]= string(runes)
		}
	}
	return strings.Join(words, " ")
}

// TruncateString truncates a string to specified length
func TruncateString(str string, maxLength int) string {
	if len(str) <= maxLength {
		return str
	}
	return str[:maxLength] + "..."
}

// SanitizeString removes unwanted characters from string
func SanitizeString(str string) string {
	// Remove leading/trailing whitespace
	str = strings.TrimSpace(str)
	
	// Replace multiple spaces with single space
	str = strings.Join(strings.Fields(str), " ")
	
	return str
}