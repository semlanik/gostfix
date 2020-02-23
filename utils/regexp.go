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
	HeaderRegExp        = "^([\x21-\x7E^:]+):(.*)"
	FoldingRegExp       = "^\\s+(.*)"
	BoundaryStartRegExp = "^--(.*)"
	BoundaryEndRegExp   = "^--(.*)--$"
	BoundaryRegExp      = "boundary=\"(.*)\""
)

const (
	UserRegExp = "^[a-zA-Z][\\w0-9\\._]*"
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
	UserChecker         *regexp.Regexp
	HeaderFinder        *regexp.Regexp
	FoldingFinder       *regexp.Regexp
	BoundaryStartFinder *regexp.Regexp
	BoundaryEndFinder   *regexp.Regexp
	BoundaryFinder      *regexp.Regexp
}

func newRegExpUtils() (*regExpUtils, error) {
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

	userChecker, err := regexp.Compile(UserRegExp)
	if err != nil {
		log.Fatalf("Invalid regexp %s\n", err)
		return nil, err
	}

	ru := &regExpUtils{
		UserChecker:         userChecker,
		HeaderFinder:        headerFinder,
		FoldingFinder:       foldingFinder,
		BoundaryStartFinder: boundaryStartFinder,
		BoundaryEndFinder:   boundaryEndFinder,
		BoundaryFinder:      boundaryFinder,
	}

	return ru, nil
}
