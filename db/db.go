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
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"log"
	"time"

	common "git.semlanik.org/semlanik/gostfix/common"
	bcrypt "golang.org/x/crypto/bcrypt"

	bson "go.mongodb.org/mongo-driver/bson"
	mongo "go.mongodb.org/mongo-driver/mongo"
	options "go.mongodb.org/mongo-driver/mongo/options"

	config "git.semlanik.org/semlanik/gostfix/config"
)

type Storage struct {
	db                  *mongo.Database
	usersCollection     *mongo.Collection
	tokensCollection    *mongo.Collection
	emailsCollection    *mongo.Collection
	allEmailsCollection *mongo.Collection
	mailsCollection     *mongo.Collection
}

func qualifiedMailCollection(user string) string {
	sum := sha1.Sum([]byte(user))
	return "mb" + hex.EncodeToString(sum[:])
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

	db := client.Database("gostfix")

	index := mongo.IndexModel{
		Keys: bson.M{
			"user": 1,
		},
		Options: options.Index().SetUnique(true),
	}

	s = &Storage{
		db:                  db,
		usersCollection:     db.Collection("users"),
		tokensCollection:    db.Collection("tokens"),
		emailsCollection:    db.Collection("emails"),
		allEmailsCollection: db.Collection("allEmails"),
		mailsCollection:     db.Collection("mails"),
	}

	//Initial database setup
	s.usersCollection.Indexes().CreateOne(context.Background(), index)
	s.tokensCollection.Indexes().CreateOne(context.Background(), index)
	s.emailsCollection.Indexes().CreateOne(context.Background(), index)

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
	if err != nil {
		return err
	}

	err = s.addEmail(user, user, true)
	if err != nil {
		s.usersCollection.DeleteOne(context.Background(), bson.M{"user": user})
		return err
	}

	//TODO: Update postfix virtual map here
	return nil
}

func (s *Storage) AddEmail(user string, email string) error {
	return s.addEmail(user, email, false)
}

func (s *Storage) addEmail(user string, email string, upsert bool) error {
	result := struct {
		User string
	}{}
	err := s.usersCollection.FindOne(context.Background(), bson.M{"user": user}).Decode(&result)

	if err != nil {
		return err
	}

	emails, err := s.GetAllEmails()

	if err != nil {
		return err
	}

	for _, existingEmail := range emails {
		if existingEmail == email {
			return errors.New("Email exists")
		}
	}

	_, err = s.emailsCollection.UpdateOne(context.Background(),
		bson.M{"user": user},
		bson.M{"$addToSet": bson.M{"email": email}},
		options.Update().SetUpsert(upsert))

	//TODO: Update postfix virtual map here
	return err
}

func (s *Storage) RemoveEmail(user string, email string) error {

	_, err := s.emailsCollection.UpdateOne(context.Background(),
		bson.M{"user": user},
		bson.M{"$pull": bson.M{"email": email}})

	//TODO: Update postfix virtual map here
	return err
}

func (s *Storage) CheckUser(user, password string) error {
	log.Printf("Check user: %s %s", user, password)
	result := struct {
		User     string
		Password string
	}{}
	err := s.usersCollection.FindOne(context.Background(), bson.M{"user": user}).Decode(&result)
	if err != nil {
		return errors.New("Invalid user or password")
	}

	if bcrypt.CompareHashAndPassword([]byte(result.Password), []byte(password)) != nil {
		return errors.New("Invalid user or password")
	}
	return nil
}

func (s *Storage) AddToken(user, token string) error {
	log.Printf("add token: %s, %s", user, token)
	s.tokensCollection.UpdateOne(context.Background(),
		bson.M{"user": user},
		bson.M{
			"$addToSet": bson.M{
				"token": bson.M{
					"token":  token,
					"expire": time.Now().Add(time.Hour * 96).Unix(),
				},
			},
		},
		options.Update().SetUpsert(true))
	return nil
}

func (s *Storage) CheckToken(user, token string) error {
	log.Printf("Check token: %s %s", user, token)
	if token == "" {
		return errors.New("Invalid token")
	}

	cur, err := s.tokensCollection.Aggregate(context.Background(),
		bson.A{
			bson.M{"$match": bson.M{"user": user}},
			bson.M{"$unwind": "$token"},
			bson.M{"$match": bson.M{"token.token": token}},
			bson.M{"$project": bson.M{"_id": 0, "token.expire": 1}},
		})

	if err != nil {
		log.Fatalln(err)
		return err
	}

	defer cur.Close(context.Background())
	if cur.Next(context.Background()) {
		result := struct {
			Token struct {
				Expire int64
			}
		}{}

		err = cur.Decode(&result)

		if err == nil && result.Token.Expire >= time.Now().Unix() {
			log.Printf("Check token %s expire: %d", user, result.Token.Expire)
			return nil
		}
	}

	return errors.New("Token expired")
}

func (s *Storage) SaveMail(email, folder string, m *common.Mail) error {
	result := &struct {
		User string
	}{}

	s.emailsCollection.FindOne(context.Background(), bson.M{"email": email}).Decode(result)

	mailsCollection := s.db.Collection(qualifiedMailCollection(result.User))
	mailsCollection.InsertOne(context.Background(), &struct {
		Email  string
		Mail   *common.Mail
		Folder string
		Read   bool
	}{
		Email:  email,
		Mail:   m,
		Folder: folder,
		Read:   false,
	}, options.InsertOne().SetBypassDocumentValidation(true))
	return nil
}

func (s *Storage) RemoveMail(user string, m *common.Mail) error {
	return nil
}

func (s *Storage) MailList(user, email, folder string) ([]*common.MailMetadata, error) {
	mailsCollection := s.db.Collection(qualifiedMailCollection(user))
	cur, err := mailsCollection.Find(context.Background(), bson.M{"email": email})

	if err != nil {
		return nil, err
	}

	var headers []*common.MailMetadata
	for cur.Next(context.Background()) {
		result := &common.MailMetadata{}
		err = cur.Decode(result)
		if err != nil {
			log.Printf("Unable to read database mail record: %s", err)
			continue
		}
		log.Printf("Add message: %s", result.Id)
		headers = append(headers, result)
	}

	log.Printf("Mails read from database: %v", headers)
	return headers, nil
}

func (s *Storage) GetMail(user string, header *common.MailHeader) (m *common.Mail, err error) {
	return nil, nil
}

func (s *Storage) GetAttachment(user string, attachmentId string) (filePath string, err error) {
	return "", nil
}

func (s *Storage) GetUsers() (users []string, err error) {
	return nil, nil
}

func (s *Storage) GetEmails(user []string) (emails []string, err error) {
	return nil, nil
}

func (s *Storage) GetAllEmails() (emails []string, err error) {
	cur, err := s.allEmailsCollection.Find(context.Background(), bson.M{})
	if cur.Next(context.Background()) {
		result := struct {
			Emails []string
		}{}
		err = cur.Decode(&result)
		if err == nil {
			return result.Emails, nil
		}
	}
	return nil, err
}
