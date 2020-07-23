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

package config

import (
	"log"
	"strings"
	"sync"

	utils "git.semlanik.org/semlanik/gostfix/utils"
	ini "gopkg.in/go-ini/ini.v1"
)

const configPath = "data/main.ini"

const (
	KeyWebPort             = "web_port"
	KeySASLPort            = "sasl_port"
	KeyPostfixConfig       = "postfix_config"
	KeyMongoAddress        = "mongo_address"
	KeyMongoUser           = "mongo_user"
	KeyMongoPassword       = "mongo_password"
	KeyAttachmentsPath     = "attachments_path"
	KeyAttachmentsUser     = "attachments_user"
	KeyAttachmentsPassword = "attachments_password"
	KeyRegistrationEnabled = "registration_enabled"
)

const (
	PostfixKeyMyDomain              = "mydomain"
	PostfixKeyVirtualMailboxMaps    = "virtual_mailbox_maps"
	PostfixKeyVirtualMailboxBase    = "virtual_mailbox_base"
	PostfixKeyVirtualMailboxDomains = "virtual_mailbox_domains"
)

type GostfixConfig gostfixConfig

var (
	once     sync.Once
	instance *gostfixConfig
)

func ConfigInstance() *GostfixConfig {

	once.Do(func() {
		instance, _ = newConfig()
	})

	return (*GostfixConfig)(instance)
}

type gostfixConfig struct {
	WebPort             string
	SASLPort            string
	MyDomain            string
	VMailboxMaps        string
	VMailboxBase        string
	VMailboxDomains     []string
	MongoUser           string
	MongoPassword       string
	MongoAddress        string
	AttachmentsPath     string
	RegistrationEnabled bool
}

func newConfig() (config *gostfixConfig, err error) {

	cfg, err := ini.Load(configPath)
	if err != nil {
		log.Fatalf("Unable to load %s\n", configPath)
		return
	}

	postfixConfigPath := cfg.Section("").Key(KeyPostfixConfig).String()
	if !utils.FileExists(postfixConfigPath) {
		log.Fatalf("Unable to find postfix config %s\n", postfixConfigPath)
		return
	}

	postfixCfg, err := ini.Load(postfixConfigPath)

	if err != nil {
		log.Fatalf("Unable to load %s: %s\n", postfixConfigPath, err)
		return
	}

	baseDir := postfixCfg.Section("").Key(PostfixKeyVirtualMailboxBase).String()

	if !utils.DirectoryExists(baseDir) {
		log.Fatalf("Base dir %s doesn't exist, postfix is not configured proper way, check %s in %s\n", baseDir, PostfixKeyVirtualMailboxBase, postfixConfigPath)
		return
	}

	maps := postfixCfg.Section("").Key(PostfixKeyVirtualMailboxMaps).String()
	mapsList := strings.Split(maps, ":")

	if len(mapsList) != 2 || mapsList[0] != "hash" {
		log.Fatalf("%s is not set proper way in %s. Should be hash:<path/to/virtualmailbox/map>, but %s provided\n", PostfixKeyVirtualMailboxMaps, postfixConfigPath, maps)
		return
	}

	if !utils.FileExists(mapsList[1] + ".db") {
		log.Fatalf("Virtual mailbox map %s doesn't exist, postfix is not configured proper way, check %s in %s\n", mapsList[1], PostfixKeyVirtualMailboxMaps, postfixConfigPath)
		return
	}

	domains := postfixCfg.Section("").Key(PostfixKeyVirtualMailboxDomains).String()
	domainsList := strings.Split(domains, " ")
	var validDomains []string
	for _, domain := range domainsList {
		if utils.RegExpUtilsInstance().DomainChecker.MatchString(domain) {
			validDomains = append(validDomains, domain)
		}
	}

	if len(validDomains) <= 0 {
		log.Fatalf("Virtual mailbox domains %s are not configured proper way, check %s in %s\n", domains, PostfixKeyVirtualMailboxDomains, postfixConfigPath)
		return
	}

	myDomain := postfixCfg.Section("").Key(PostfixKeyMyDomain).String()

	if len(myDomain) <= 0 {
		myDomain = "localhost"
	}

	mongoUser := cfg.Section("").Key(KeyMongoUser).String()

	mongoPassword := cfg.Section("").Key(KeyMongoPassword).String()

	mongoAddress := cfg.Section("").Key(KeyMongoAddress).String()

	if mongoAddress == "" {
		mongoAddress = "localhost:27017"
	}

	attachmentsPath := cfg.Section("").Key(KeyAttachmentsPath).String()

	if attachmentsPath == "" {
		attachmentsPath = "attachments"
	}

	registrationEnabled := cfg.Section("").Key(KeyRegistrationEnabled).String()

	webPort := cfg.Section("").Key(KeyWebPort).String()
	if webPort == "" {
		webPort = "65200"
	}

	saslPort := cfg.Section("").Key(KeySASLPort).String()
	if saslPort == "" {
		saslPort = "65201"
	}

	config = &gostfixConfig{
		WebPort:             webPort,
		SASLPort:            saslPort,
		MyDomain:            myDomain,
		VMailboxBase:        baseDir,
		VMailboxMaps:        mapsList[1] + ".db",
		VMailboxDomains:     validDomains,
		MongoUser:           mongoUser,
		MongoPassword:       mongoPassword,
		MongoAddress:        mongoAddress,
		AttachmentsPath:     attachmentsPath,
		RegistrationEnabled: registrationEnabled == "true",
	}
	return
}
