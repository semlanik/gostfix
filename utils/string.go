/*
 * MIT License
 *
 * Copyright (c) 2020 Alexey Edelev <semlanik@gmail.com>
 *
 * This file is part of gostfix project https://git.semlanik.org/semlanik/gostfix
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy of this
 * software and associated documentation files (the "Software"), to deal in the Software
 * without restriction, including without limitation the rights to use, copy, modify,
 * merge, publish, distribute, sublicense, and/or sell copies of the Software, and
 * to permit persons to whom the Software is furnished to do so, subject to the following
 * conditions:
 *
 * The above copyright notice and this permission notice shall be included in all copies
 * or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED,
 * INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR
 * PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE
 * FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR
 * OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER
 * DEALINGS IN THE SOFTWARE.
 */

package utils

import (
	"regexp"
	"strings"
)

func StartsWith(s, key string) bool {
	return strings.Index(s, key) == 0
}

func RemoveSubString(text *string, begin string, end string) {
	headIndex := strings.Index(*text, begin)
	if headIndex >= 0 {
		headEndIndex := strings.Index(*text, end)
		runes := []rune(*text)
		runes = append(runes[0:headIndex], runes[headEndIndex+len(end):]...)
		*text = string(runes)
	}
}

func SanitizeTags(text *string) {
	re := regexp.MustCompile(`</?html[^<>]*>`)
	*text = string(re.ReplaceAll([]byte(*text), []byte{}))

	re = regexp.MustCompile(`</?body[^<>]*>`)
	*text = string(re.ReplaceAll([]byte(*text), []byte{}))

	RemoveSubString(text, "<head", "/head>")
	RemoveSubString(text, "<style", "/style>")
}
