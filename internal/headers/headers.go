// Package headers used to process headers for http requests and responses
package headers

import (
	"bytes"
	"fmt"
	"strings"
	"unicode"
)


type Headers map[string]string

const crlf = "\r\n"

func NewHeaders() Headers {
	return Headers{}
}

var validRFCSymbols = map[string]struct{}{
	"!": {}, 	
	"#": {},
	"$": {},
	"%": {}, 
	"&": {}, 
	"'": {}, 
	"*": {}, 
	"+": {}, 
	"-": {}, 
	".": {}, 
	"^": {}, 
	"_": {}, 
	"`": {}, 
	"|": {}, 
	"~": {},
}

func containsOnlyValidTokens(fieldName string) bool {
	for _, c := range fieldName {
		if _, ok := validRFCSymbols[string(c)]; !unicode.IsDigit(rune(c)) && !unicode.IsLetter(rune(c)) && !ok {
			return false
		}
	}
	return true
}

func (h Headers) Parse(data []byte) (n int, done bool, err error){
	idx := bytes.Index(data, []byte(crlf))
	if idx == -1 {
		return 0, false, nil
	} 
	
	fmt.Println(string(data[:idx]))
	if idx == 0 {
		return len(crlf), true, nil
	}
	
	colonIndex := bytes.Index(data, []byte(":"))
	if colonIndex == -1 {
		return 0, false, fmt.Errorf("invalid field line syntax")
	}
	splitData := []string{string(data[:colonIndex]), string(data[colonIndex + 1:idx])}

	if unicode.IsSpace(rune(splitData[0][len(splitData[0]) - 1])) {
		return 0, false, fmt.Errorf("invalid field line syntax")
	}
	
	unparsedKey := strings.Fields(splitData[0])
	unparsedValue := strings.Fields(splitData[1])
	
	if len(unparsedKey) > 1 || len(unparsedValue) > 1 {
		return 0, false,  fmt.Errorf("invalid field line syntax") 
	}
	
	key := unparsedKey[0]
	value := unparsedValue[0]
	
	key = strings.ToLower(key)
	if !containsOnlyValidTokens(key) || len(key) <= 1 {
		return 0, false, fmt.Errorf("invalid field name")
	}
	
	fmt.Println("key value", key, value)

	if v, ok := h[key]; ok{
		h[key] = v + "," + value
	} else {
		h[key] = value
	}

	return idx + len(crlf), false, nil 
}


