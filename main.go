package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
	"unicode"
)

const (
	endSentencePattern = `([\w ,.!?]+)?`
	vowel              = "[aiueo]"
	vowelNoE           = "[aiuo]"
	vowelNoIE          = "[auo]"
	zackqyWord         = "[jzckq]"
)

func main() {
	// Create a scanner that looks at standard input
	scanner := bufio.NewScanner(os.Stdin)

	// This loop will block and wait for input, running infinitely
	for scanner.Scan() {
		input := scanner.Text()
		fmt.Printf("%s", Owowify(input))
	}

	// Check if any error occurred during scanning
	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "reading standard input:", err)
	}
}

func Owowify(text string) string {
	// OwO emote translations
	text = subOwoEmote(text, `(?i)(i(?:'|)m(?:\s+|\s+so+\s+)bored)`+endSentencePattern, "-w-")
	text = subOwoEmote(text, `(?i)(love\s+(?:you|him|her|them))`+endSentencePattern, "uwu")
	text = subOwoEmote(text, `(?i)(i\s+don(?:'|)t\s+care|i\s*d\s*c)`+endSentencePattern, "0w0")

	// world substitution: l[ou]ve? -> luv
	reLove := regexp.MustCompile(`(?i)l[ou]ve?`)
	text = reLove.ReplaceAllStringFunc(text, func(m string) string {
		return subSameCase(m, "luv")
	})

	// OwO translation: r -> w, unless r is alone
	reR1 := regexp.MustCompile(`(\w)[rR]`)
	text = reR1.ReplaceAllStringFunc(text, func(m string) string {
		// m is 2 characters long: some word char + r/R
		return m[:1] + subSameCase(m[1:], "w")
	})
	reR2 := regexp.MustCompile(`[rR](\w)`)
	text = reR2.ReplaceAllStringFunc(text, func(m string) string {
		// m is 2 characters long: r/R + some word char
		return subSameCase(m[:1], "w") + m[1:]
	})

	// l -> w logic (replacing lookarounds with matching contexts)
	// JavaScript regex: /(?<!([wl]${vowel}*))(?:l(?=\w)|(?<=\w)l)(?!([wl]))/gi
	// We handle this using a function match since Go lacks lookarounds.
	reL := regexp.MustCompile(`(?i)[a-z]+`)
	text = reL.ReplaceAllStringFunc(text, func(word string) string {
		return replaceLInWord(word)
	})

	// na -> nya, nu -> nyu, no -> nyo, ne -> nye
	reN := regexp.MustCompile(`(?i)n` + vowelNoE + `+`)
	text = reN.ReplaceAllStringFunc(text, func(m string) string {
		v := m[1:]
		return subSameCase(m, "ny"+v)
	})

	// ma -> mya, mu -> myu, mo -> myo
	// JS checks for negative lookahead (?!w*${zackqyWord})
	reM := regexp.MustCompile(`(?i)m` + vowelNoIE + `+`)
	text = replaceWithLookaheadCheck(text, reM, "my", zackqyWord)

	// pa -> pwa, pu -> pwu, po -> pwo
	reP := regexp.MustCompile(`(?i)p` + vowelNoIE + `+`)
	text = replaceWithLookaheadCheck(text, reP, "pw", zackqyWord)

	return text
}

// Emulates the subOwoEmote logic
func subOwoEmote(text, pattern, emote string) string {
	re := regexp.MustCompile(pattern)
	matchEndSpace := regexp.MustCompile(`^\s+$`)

	return re.ReplaceAllStringFunc(text, func(m string) string {
		submatches := re.FindStringSubmatch(m)
		if len(submatches) < 3 {
			return m
		}
		sentenceBeforeEnd := submatches[1]
		endSentence := submatches[2]

		if endSentence == "" || matchEndSpace.MatchString(endSentence) {
			return sentenceBeforeEnd + " " + emote
		}
		return m
	})
}

// Replaces components of 'l' and 'L' based on the specific surrounding character rules
func replaceLInWord(word string) string {
	runes := []rune(word)
	reVowels := regexp.MustCompile(`(?i)^[wl]` + vowel + `*$`)

	for i := 0; i < len(runes); i++ {
		if unicode.ToLower(runes[i]) != 'l' {
			continue
		}
		// Must match (l(?=\w) or (?<=\w)l) -> basically 'l' cannot be an isolated single letter
		if len(runes) == 1 {
			continue
		}
		// (?!([wl])) -> next char cannot be w or l
		if i+1 < len(runes) && (unicode.ToLower(runes[i+1]) == 'w' || unicode.ToLower(runes[i+1]) == 'l') {
			continue
		}
		// (?<!([wl]${vowel}*)) -> backward check within the word prefix
		prefix := string(runes[:i])
		if len(prefix) > 0 && reVowels.MatchString(prefix) {
			continue
		}

		// Apply replacement
		if unicode.IsUpper(runes[i]) {
			runes[i] = 'W'
		} else {
			runes[i] = 'w'
		}
	}
	return string(runes)
}

// Mimics negative lookahead checks for M and P rules
func replaceWithLookaheadCheck(text string, re *regexp.Regexp, insert string, avoid string) string {
	reAvoid := regexp.MustCompile(`(?i)^w*` + avoid)

	return re.ReplaceAllStringFunc(text, func(m string) string {
		idx := strings.Index(text, m)
		if idx != -1 && idx+len(m) < len(text) {
			following := text[idx+len(m):]
			if reAvoid.MatchString(following) {
				return m // Skip if lookahead matches forbidden structure
			}
		}
		v := m[1:]
		return subSameCase(m, insert+v)
	})
}

// Keeps character casing synchronous between origin and replacement slices
func subSameCase(inputText, replaceText string) string {
	inRunes := []rune(inputText)
	repRunes := []rune(replaceText)
	var result strings.Builder

	for i := 0; i < len(repRunes); i++ {
		if i < len(inRunes) {
			if unicode.IsUpper(inRunes[i]) {
				result.WriteRune(unicode.ToUpper(repRunes[i]))
			} else if unicode.IsLower(inRunes[i]) {
				result.WriteRune(unicode.ToLower(repRunes[i]))
			} else {
				result.WriteRune(repRunes[i])
			}
		} else {
			result.WriteRune(repRunes[i])
		}
	}
	return result.String()
}
