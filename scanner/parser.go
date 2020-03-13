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
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"git.semlanik.org/semlanik/gostfix/common"
	"git.semlanik.org/semlanik/gostfix/config"
	utils "git.semlanik.org/semlanik/gostfix/utils"
	"github.com/google/uuid"
	enmime "github.com/jhillyerd/enmime"
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
	log.Println("Parse file")
	defer log.Println("Exit parse")

	var emails []*common.Mail

	pd := &parseData{}
	pd.reset()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		log.Println("Scan next line")

		currentText := scanner.Text()
		if utils.RegExpUtilsInstance().MailIndicator.MatchString(currentText) {
			if pd.mandatoryHeaders == AllHeaderMask {
				pd.parseBody()
				emails = append(emails, pd.email)
			}
			pd.reset()
			fmt.Println("Found new email" + currentText)
			continue
		}

		switch pd.state {
		case StateHeaderScan:
			if currentText == "" {
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
				pd.parseHeader(currentText)
			}
		case StateBodyScan:
			// if currentText == "" {
			// 	if pd.state == StateBodyScan && pd.activeBoundary == "" {
			// 		if pd.mandatoryHeaders == AllHeaderMask {
			// 			pd.parseBody()
			// 			emails = append(emails, pd.email)
			// 		}
			// 		pd.reset()
			// 		continue
			// 	}
			// }

			// if pd.activeBoundary != "" {
			pd.bodyData += currentText + "\n"
			capture := utils.RegExpUtilsInstance().BoundaryEndFinder.FindStringSubmatch(currentText)
			if len(capture) == 2 && pd.activeBoundary == capture[1] {
				pd.state = StateBodyScan
				pd.activeBoundary = ""
			}
			// }
		}
	}

	if pd.state == StateBodyScan {
		if pd.mandatoryHeaders == AllHeaderMask {
			pd.parseBody()
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
		case "x-original-to":
			if pd.email.Header.To == "" {
				pd.previousHeader = &pd.email.Header.To
				pd.mandatoryHeaders |= ToHeaderMask
			}
		case "cc":
			pd.previousHeader = &pd.email.Header.Cc
		case "bcc":
			pd.previousHeader = &pd.email.Header.Bcc
			pd.mandatoryHeaders |= ToHeaderMask
		case "subject":
			pd.previousHeader = &pd.email.Header.Subject
		case "date":
			pd.previousHeader = nil
			unixTime, err := parseDate(strings.Trim(capture[2], " \t"))
			if err == nil {
				pd.email.Header.Date = unixTime
				pd.mandatoryHeaders |= DateHeaderMask
			} else {
				log.Printf("Unable to parse message: %s\n", err)
			}
		case "content-type":
			pd.previousHeader = &pd.bodyContentType
		default:
			pd.previousHeader = nil
		}

		if pd.previousHeader != nil {
			*pd.previousHeader = strings.Trim(capture[2], " \t")
		}
		return
	}

	//Parse folding
	capture = utils.RegExpUtilsInstance().FoldingFinder.FindStringSubmatch(headerRaw)
	if len(capture) == 2 && pd.previousHeader != nil {
		*pd.previousHeader += capture[1]
	}
}

func (pd *parseData) parseBody() {
	buffer := bytes.NewBufferString("content-type:" + pd.bodyContentType + "\n\n" + pd.bodyData)
	en, err := enmime.ReadEnvelope(buffer)
	if err != nil {
		log.Printf("Unable to read mail body %s\n\nBody content: %s\n\n", err, pd.bodyData)
		return
	}

	pd.email.Body = &common.MailBody{}

	pd.email.Body.PlainText = en.Text
	pd.email.Body.RichText = en.HTML

	for _, attachment := range en.Attachments {
		uuid := uuid.New()
		fileName := hex.EncodeToString(uuid[:])
		attachmentFile, err := os.Create(config.ConfigInstance().AttachmentsPath + "/" + fileName)
		log.Printf("Attachment found %s\n", fileName)
		if err != nil {
			log.Printf("Unable to save attachment %s %s\n", fileName, err)
			continue
		}
		pd.email.Body.Attachments = append(pd.email.Body.Attachments, &common.AttachmentHeader{
			Id:          fileName,
			FileName:    attachment.FileName,
			ContentType: attachment.ContentType,
		})
		attachmentFile.Write(attachment.Content)
	}
}

func parseDate(stringDate string) (int64, error) {
	formatsToTest := []string{
		"Mon, _2 Jan 2006 15:04:05 -0700",
		time.RFC1123Z,
		time.RFC1123,
		time.UnixDate,
		"Mon,  _2 Jan 2006 15:04:05 -0700 (MST)",
		"Mon, _2 Jan 2006 15:04:05 -0700 (MST)"}
	var err error
	for _, format := range formatsToTest {
		dateTime, err := time.Parse(format, stringDate)
		if err == nil {
			return dateTime.Unix(), nil
		}
	}

	return 0, errors.New("Invalid date format " + stringDate + " , " + err.Error())
}
