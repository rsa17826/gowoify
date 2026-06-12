package main

import (
	"fmt"
	"io"
	"os"
	"strings"
	"unicode"

	"github.com/dlclark/regexp2"
)

const (
	endSentencePattern = `([\w ,.!?]+)?`
	vowel              = "[aiueo]"
	vowelNoE           = "[aiuo]"
	vowelNoIE          = "[auo]"
	zackqyWord         = "[jzckq]"
)

func main() {
	inputBytes, err := io.ReadAll(os.Stdin)
	if err != nil {
		fmt.Fprintln(os.Stderr, "reading standard input:", err)
		return
	}

	input := string(inputBytes)

	// Process and print everything with original newlines intact
	fmt.Print(Owowify(input))
}

func Owowify(text string) string {
	endSentencePattern := `([\w ,.!?]+)?`
	vowel := "[aiueo]"
	vowelNoE := "[aiuo]"
	vowelNoIE := "[auo]"
	zackqyWord := "[jzckq]"

	// 1. OwO Emotes
	text = replaceAllStringFuncBuf(
		regexp2.MustCompile(`(?i)(i(?:'|)m(?:\s+|\s+so+\s+)bored)`+endSentencePattern, 0),
		text,
		subOwoEmote("-w-"),
	)
	text = replaceAllStringFuncBuf(
		regexp2.MustCompile(`(?i)(love\s+(?:you|him|her|them))`+endSentencePattern, 0),
		text,
		subOwoEmote("uwu"),
	)
	text = replaceAllStringFuncBuf(
		regexp2.MustCompile(`(?i)(i\s+don(?:'|)t\s+care|i\s*d\s*c)`+endSentencePattern, 0),
		text,
		subOwoEmote("0w0"),
	)

	// 2. Word substitution
	text = replaceAllStringFuncBuf(
		regexp2.MustCompile(`(?i)l[ou]ve?`, 0),
		text,
		func(m regexp2.Match) string {
			return subSameCase(m.String(), "luv")
		},
	)

	// 3. OwO translation
	// r -> w (unless r is alone)
	text = replaceAllStringFuncBuf(
		regexp2.MustCompile(`(?i)(?<=\w)r`, 0),
		text,
		func(m regexp2.Match) string {
			return subSameCase(m.String(), "w")
		},
	)
	text = replaceAllStringFuncBuf(
		regexp2.MustCompile(`(?i)r(?=\w)`, 0),
		text,
		func(m regexp2.Match) string {
			return subSameCase(m.String(), "w")
		},
	)

	// l -> w adjustments
	lPattern := fmt.Sprintf(`(?i)(?<!([wl]%s*))(?:l(?=\w)|(?<=\w)l)(?!([wl]))`, vowel)
	text = replaceAllStringFuncBuf(
		regexp2.MustCompile(lPattern, 0),
		text,
		func(m regexp2.Match) string {
			return subSameCase(m.String(), "w")
		},
	)

	// n -> ny variants
	nPattern1 := fmt.Sprintf(`[nN](%s+)`, vowelNoE)
	text = replaceAllStringFuncBuf(
		regexp2.MustCompile(nPattern1, 0),
		text,
		func(m regexp2.Match) string {
			v := m.Groups()[1].Captures[0].String()
			return subSameCase(m.String(), "ny"+v)
		},
	)
	nPattern2 := fmt.Sprintf(`N(%s+)`, strings.ToUpper(vowelNoE))
	text = replaceAllStringFuncBuf(
		regexp2.MustCompile(nPattern2, 0),
		text,
		func(m regexp2.Match) string {
			v := m.Groups()[1].Captures[0].String()
			return subSameCase(m.String(), "ny"+v)
		},
	)

	// m -> my variants
	mPattern1 := fmt.Sprintf(`[mM](%s+)(?!w*%s)`, vowelNoIE, zackqyWord)
	text = replaceAllStringFuncBuf(
		regexp2.MustCompile(mPattern1, 0),
		text,
		func(m regexp2.Match) string {
			v := m.Groups()[1].Captures[0].String()
			return subSameCase(m.String(), "my"+v)
		},
	)
	mPattern2 := fmt.Sprintf(`M(%s+)(?!w*%s)`, strings.ToUpper(vowelNoE), zackqyWord)
	text = replaceAllStringFuncBuf(
		regexp2.MustCompile(mPattern2, 0),
		text,
		func(m regexp2.Match) string {
			v := m.Groups()[1].Captures[0].String()
			return subSameCase(m.String(), "my"+v)
		},
	)

	// p -> pw variants
	pPattern1 := fmt.Sprintf(`[pP](%s+)(?!w*%s)`, vowelNoIE, zackqyWord)
	text = replaceAllStringFuncBuf(
		regexp2.MustCompile(pPattern1, 0),
		text,
		func(m regexp2.Match) string {
			v := m.Groups()[1].Captures[0].String()
			return subSameCase(m.String(), "pw"+v)
		},
	)
	pPattern2 := fmt.Sprintf(`P(%s+)(?!w*%s)`, strings.ToUpper(vowelNoIE), zackqyWord)
	text = replaceAllStringFuncBuf(
		regexp2.MustCompile(pPattern2, 0),
		text,
		func(m regexp2.Match) string {
			v := m.Groups()[1].Captures[0].String()
			return subSameCase(m.String(), "pw"+v)
		},
	)

	return text
}

// subOwoEmote replicates the JS closure for replacing end sentences with emotes
func subOwoEmote(emote string) func(regexp2.Match) string {
	matchEndSpace := regexp2.MustCompile(`^\s+$`, 0)

	return func(m regexp2.Match) string {
		g := m.Groups()
		sentenceBeforeEnd := g[1].Captures[0].String()

		var endSentence string
		if len(g) > 2 && len(g[2].Captures) > 0 {
			endSentence = g[2].Captures[0].String()
		}

		isSpace, _ := matchEndSpace.MatchString(endSentence)
		if endSentence == "" || isSpace {
			return fmt.Sprintf("%s %s", sentenceBeforeEnd, emote)
		}
		return m.String()
	}
}

// subSameCase preserves upper/lower casing based on input template
func subSameCase(inputText, replaceText string) string {
	var result strings.Builder
	inputRunes := []rune(inputText)
	replaceRunes := []rune(replaceText)

	for i := 0; i < len(replaceRunes); i++ {
		if i < len(inputRunes) {
			if unicode.IsUpper(inputRunes[i]) {
				result.WriteRune(unicode.ToUpper(replaceRunes[i]))
			} else if unicode.IsLower(inputRunes[i]) {
				result.WriteRune(unicode.ToLower(replaceRunes[i]))
			} else {
				result.WriteRune(replaceRunes[i])
			}
		} else {
			result.WriteRune(replaceRunes[i])
		}
	}
	return result.String()
}

// Helper function to safely evaluate a string replacement function over all matches.
// m.Index and m.Length are rune offsets (not byte offsets), so we convert to []rune
// first to avoid slicing through multi-byte UTF-8 sequences (e.g. box-drawing chars).
func replaceAllStringFuncBuf(re *regexp2.Regexp, input string, replacer func(regexp2.Match) string) string {
	runes := []rune(input)
	var result strings.Builder
	lastIndex := 0

	m, err := re.FindStringMatch(input)
	for err == nil && m != nil {
		result.WriteString(string(runes[lastIndex:m.Index]))
		result.WriteString(replacer(*m))
		lastIndex = m.Index + m.Length
		m, err = re.FindNextMatch(m)
	}
	result.WriteString(string(runes[lastIndex:]))
	return result.String()
}
