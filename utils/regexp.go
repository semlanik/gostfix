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
	"log"
	"regexp"
	"sync"
)

const (
	NewMailIndicator    = "^(From\\s).*"
	HeaderRegExp        = "^([\x21-\x7E^:]+):(.*)"
	FoldingRegExp       = "^\\s+(.*)"
	BoundaryStartRegExp = "^--(.*)"
	BoundaryEndRegExp   = "^--(.*)--$"
	BoundaryRegExp      = "boundary=\"(.*)\""
	MailboxRegExp       = "^/m(\\d+)/?(.*)"
	FullNameRegExp      = "^[\\w]+[\\w ]*$"
)

const (
	DomainRegExp = "(?:[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?\\.)+[a-z0-9][a-z0-9-]{0,61}[a-z0-9]"
	EmailRegExp  = "(?:[a-z0-9!#$%&'*+/=?^_`{|}~-]+(?:\\.[a-z0-9!#$%&'*+/=?^_`{|}~-]+)*|\"(?:[\x01-\x08\x0b\x0c\x0e-\x1f\x21\x23-\x5b\x5d-\x7f]|\\[\x01-\x09\x0b\x0c\x0e-\x7f])*\")@(?:(?:[a-z0-9](?:[a-z0-9-]*[a-z0-9])?\\.)+[a-z0-9](?:[a-z0-9-]*[a-z0-9])?|\\[(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?|[a-z0-9-]*[a-z0-9]:(?:[\x01-\x08\x0b\x0c\x0e-\x1f\x21-\x5a\x53-\x7f]|\\[\x01-\x09\x0b\x0c\x0e-\x7f])+)\\])"
)

type RegExpUtils regExpUtils

var (
	once     sync.Once
	instance *regExpUtils
)

func RegExpUtilsInstance() *RegExpUtils {

	once.Do(func() {
		instance, _ = newRegExpUtils()
	})

	return (*RegExpUtils)(instance)
}

type regExpUtils struct {
	MailIndicator       *regexp.Regexp
	DomainChecker       *regexp.Regexp
	EmailChecker        *regexp.Regexp
	HeaderFinder        *regexp.Regexp
	FoldingFinder       *regexp.Regexp
	BoundaryStartFinder *regexp.Regexp
	BoundaryEndFinder   *regexp.Regexp
	BoundaryFinder      *regexp.Regexp
	MailboxFinder       *regexp.Regexp
	FullNameChecker     *regexp.Regexp
}

func newRegExpUtils() (*regExpUtils, error) {
	mailIndicator, err := regexp.Compile(NewMailIndicator)
	if err != nil {
		log.Fatalf("Invalid regexp %s\n", err)
		return nil, err
	}

	headerFinder, err := regexp.Compile(HeaderRegExp)
	if err != nil {
		log.Fatalf("Invalid regexp %s\n", err)
		return nil, err
	}

	foldingFinder, err := regexp.Compile(FoldingRegExp)
	if err != nil {
		log.Fatalf("Invalid regexp %s\n", err)
		return nil, err
	}

	boundaryStartFinder, err := regexp.Compile(BoundaryStartRegExp)
	if err != nil {
		log.Fatalf("Invalid regexp %s\n", err)
		return nil, err
	}

	boundaryEndFinder, err := regexp.Compile(BoundaryEndRegExp)
	if err != nil {
		log.Fatalf("Invalid regexp %s\n", err)
		return nil, err
	}

	boundaryFinder, err := regexp.Compile(BoundaryRegExp)
	if err != nil {
		log.Fatalf("Invalid regexp %s\n", err)
		return nil, err
	}

	domainChecker, err := regexp.Compile(DomainRegExp)
	if err != nil {
		log.Fatalf("Invalid regexp %s\n", err)
		return nil, err
	}

	emailChecker, err := regexp.Compile(EmailRegExp)
	if err != nil {
		log.Fatalf("Invalid regexp %s\n", err)
		return nil, err
	}

	mailboxFinder, err := regexp.Compile(MailboxRegExp)
	if err != nil {
		log.Fatalf("Invalid regexp %s\n", err)
		return nil, err
	}

	fullNameChecker, err := regexp.Compile(FullNameRegExp)
	if err != nil {
		log.Fatalf("Invalid regexp %s\n", err)
		return nil, err
	}

	ru := &regExpUtils{
		MailIndicator:       mailIndicator,
		EmailChecker:        emailChecker,
		HeaderFinder:        headerFinder,
		FoldingFinder:       foldingFinder,
		BoundaryStartFinder: boundaryStartFinder,
		BoundaryEndFinder:   boundaryEndFinder,
		BoundaryFinder:      boundaryFinder,
		DomainChecker:       domainChecker,
		MailboxFinder:       mailboxFinder,
		FullNameChecker:     fullNameChecker,
	}

	return ru, nil
}
