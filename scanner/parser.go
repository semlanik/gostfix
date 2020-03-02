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

package scanner

import (
	"bufio"
	"log"
	"strings"
	"time"

	"git.semlanik.org/semlanik/gostfix/common"
	utils "git.semlanik.org/semlanik/gostfix/utils"
)

const (
	StateHeaderScan = iota
	StateBodyScan
)

const (
	AtLeastOneHeaderMask = 1 << iota
	FromHeaderMask
	DateHeaderMask
	ToHeaderMask
	AllHeaderMask = 15
)

type parseData struct {
	state            int
	mandatoryHeaders int
	previousHeader   *string
	email            *common.Mail
	bodyContentType  string
	bodyData         string
	activeBoundary   string
}

func (pd *parseData) reset() {
	*pd = parseData{
		state:            StateHeaderScan,
		previousHeader:   nil,
		mandatoryHeaders: 0,
		email:            NewEmail(),
		bodyContentType:  "plain/text",
		bodyData:         "",
		activeBoundary:   "",
	}
}

func parseFile(file *utils.LockedFile) []*common.Mail {
	var emails []*common.Mail

	pd := &parseData{}
	pd.reset()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		switch pd.state {
		case StateHeaderScan:
			if scanner.Text() == "" {
				if pd.mandatoryHeaders&AtLeastOneHeaderMask == AtLeastOneHeaderMask { //Cause we read at least one header
					pd.previousHeader = nil
					boundaryCapture := utils.RegExpUtilsInstance().BoundaryFinder.FindStringSubmatch(pd.bodyContentType)
					if len(boundaryCapture) == 2 {
						pd.activeBoundary = boundaryCapture[1]
					} else {
						pd.activeBoundary = ""
					}
					pd.state = StateBodyScan
				}
			} else {
				pd.parseHeader(scanner.Text())
			}
		case StateBodyScan:
			if scanner.Text() == "" {
				if pd.state == StateBodyScan && pd.activeBoundary == "" {
					if pd.mandatoryHeaders == AllHeaderMask {
						emails = append(emails, pd.email)
					}
					pd.reset()
					continue
				}
			}

			if pd.activeBoundary != "" {
				pd.bodyData += scanner.Text() + "\n"
				capture := utils.RegExpUtilsInstance().BoundaryEndFinder.FindStringSubmatch(scanner.Text())
				if len(capture) == 2 && pd.activeBoundary == capture[1] {
					pd.state = StateBodyScan
					pd.activeBoundary = ""
				}
			}
		}
	}

	if pd.state == StateBodyScan {
		if pd.mandatoryHeaders == AllHeaderMask {
			emails = append(emails, pd.email)
		}
		pd.reset()
	}
	return emails
}

func (pd *parseData) parseHeader(headerRaw string) {
	capture := utils.RegExpUtilsInstance().HeaderFinder.FindStringSubmatch(headerRaw)
	//Parse header
	if len(capture) == 3 {
		// fmt.Printf("capture Header %s : %s\n", strings.ToLower(capture[0]), strings.ToLower(capture[1]))
		header := strings.ToLower(capture[1])
		pd.mandatoryHeaders |= AtLeastOneHeaderMask
		switch header {
		case "from":
			pd.previousHeader = &pd.email.Header.From
			pd.mandatoryHeaders |= FromHeaderMask
		case "to":
			pd.previousHeader = &pd.email.Header.To
			pd.mandatoryHeaders |= ToHeaderMask
		case "cc":
			pd.previousHeader = &pd.email.Header.Cc
		case "bcc":
			pd.previousHeader = &pd.email.Header.Bcc
			pd.mandatoryHeaders |= ToHeaderMask
		case "subject":
			pd.previousHeader = &pd.email.Header.Subject
		case "date":
			pd.previousHeader = nil
			time, err := time.Parse(time.RFC1123Z, strings.Trim(capture[2], " \t"))
			if err == nil {
				pd.email.Header.Date = time.Unix()
				pd.mandatoryHeaders |= DateHeaderMask
			}
			log.Printf("Invalid date format %s, %s", strings.Trim(capture[2], " \t"), err)
		case "content-type":
			pd.previousHeader = &pd.bodyContentType
		default:
			pd.previousHeader = nil
		}

		if pd.previousHeader != nil {
			*pd.previousHeader = capture[2]
		}
		return
	}

	//Parse folding
	capture = utils.RegExpUtilsInstance().FoldingFinder.FindStringSubmatch(headerRaw)
	if len(capture) == 2 && pd.previousHeader != nil {
		*pd.previousHeader += capture[1]
	}
}
