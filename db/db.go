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

package db

import (
	"context"
	"time"

	common "git.semlanik.org/semlanik/gostfix/common"
	bcrypt "golang.org/x/crypto/bcrypt"

	bson "go.mongodb.org/mongo-driver/bson"
	mongo "go.mongodb.org/mongo-driver/mongo"
	options "go.mongodb.org/mongo-driver/mongo/options"

	config "git.semlanik.org/semlanik/gostfix/config"
)

type Storage struct {
	usersCollection  *mongo.Collection
	tokensCollection *mongo.Collection
}

func NewStorage() (s *Storage, err error) {
	fullUrl := "mongodb://"
	if config.ConfigInstance().MongoUser != "" {
		fullUrl += config.ConfigInstance().MongoUser
		if config.ConfigInstance().MongoPassword != "" {
			fullUrl += ":" + config.ConfigInstance().MongoPassword
		}
		fullUrl += "@"
	}

	fullUrl += config.ConfigInstance().MongoAddress

	client, err := mongo.NewClient(options.Client().ApplyURI(fullUrl))
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	err = client.Connect(ctx)
	if err != nil {
		return nil, err
	}

	s = &Storage{
		usersCollection:  client.Database("gostfix").Collection("users"),
		tokensCollection: client.Database("gostfix").Collection("tokens"),
	}
	return
}

func (s *Storage) AddUser(user, password, fullName string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	hashString := string(hash)
	userInfo := bson.M{
		"user":     user,
		"password": hashString,
		"fullName": fullName,
	}
	_, err = s.usersCollection.InsertOne(context.Background(), userInfo)
	return err
}

func (s *Storage) CheckUser(user, password string) error {
	result := struct {
		User     string
		Password string
	}{}
	err := s.usersCollection.FindOne(context.Background(), bson.M{"user": user}).Decode(&result)
	if err != nil {
		return err
	}

	return nil
}

func (s *Storage) AddToken(user, token string) error {
	return nil
}

func (s *Storage) CheckToken(user, token string) error {
	return nil
}

func (s *Storage) SaveMail(user string, m *common.Mail) error {
	return nil
}

func (s *Storage) RemoveMail(user string, m *common.Mail) error {
	return nil
}

func (s *Storage) MailList(user string) ([]*common.MailHeader, error) {
	return nil, nil
}

func (s *Storage) GetMail(user string, header *common.MailHeader) (m *common.Mail, err error) {
	return nil, nil
}

func (s *Storage) GetAttachment(user string, attachmentId string) (filePath string, err error) {
	return "", nil
}
