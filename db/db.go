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
	"fmt"
	"log"
	"time"

	common "git.semlanik.org/semlanik/gostfix/common"
	bcrypt "golang.org/x/crypto/bcrypt"

	bson "go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
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
	log.Printf("Add token: %s\n", user)
	s.tokensCollection.UpdateOne(context.Background(),
		bson.M{"user": user},
		bson.M{
			"$addToSet": bson.M{
				"token": bson.M{
					"token":  token,
					"expire": time.Now().Add(time.Hour * 24).Unix(),
				},
			},
		},
		options.Update().SetUpsert(true))
	s.CleanupTokens(user)
	return nil
}

func (s *Storage) CheckToken(user, token string) error {
	if token == "" {
		return errors.New("Invalid token")
	}

	cur, err := s.tokensCollection.Aggregate(context.Background(),
		bson.A{
			bson.M{"$match": bson.M{"user": user}},
			bson.M{"$unwind": "$token"},
			bson.M{"$match": bson.M{"token.token": token}},
		})

	if err != nil {
		log.Fatalln(err)
		return err
	}

	ok := false
	defer cur.Close(context.Background())
	if cur.Next(context.Background()) {
		result := struct {
			Token struct {
				Expire int64
			}
		}{}

		err = cur.Decode(&result)

		ok = err == nil && result.Token.Expire >= time.Now().Unix()
	}

	if ok {
		//TODO: Renew token
		return nil
	}

	return errors.New("Token expired")
}

func (s *Storage) RemoveToken(user, token string) error {
	s.CleanupTokens(user)

	_, err := s.tokensCollection.UpdateOne(context.Background(), bson.M{"user": user}, bson.M{"$pull": bson.M{"token": bson.M{"token": token}}})
	if err != nil {
		log.Printf("Unable to remove token %s", err)
	}

	return err
}

func (s *Storage) CleanupTokens(user string) {
	log.Printf("Cleanup tokens: %s\n", user)

	cur, err := s.tokensCollection.Aggregate(context.Background(),
		bson.A{
			bson.M{"$match": bson.M{"user": user}},
			bson.M{"$unwind": "$token"},
		})

	if err != nil {
		log.Fatalln(err)
	}

	type tokenMetadata struct {
		Expire int64
		Token  string
	}

	tokensToKeep := bson.A{}
	defer cur.Close(context.Background())
	for cur.Next(context.Background()) {
		result := struct {
			Token *tokenMetadata
		}{
			Token: &tokenMetadata{},
		}

		err = cur.Decode(&result)
		if err == nil && result.Token.Expire >= time.Now().Unix() {
			tokensToKeep = append(tokensToKeep, result.Token)
		} else {
			log.Printf("Expired token found for %s : %d", user, result.Token.Expire)
		}
	}

	_, err = s.tokensCollection.UpdateOne(context.Background(), bson.M{"user": user}, bson.M{"$set": bson.M{"token": tokensToKeep}})
	return
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
		Trash  bool
	}{
		Email:  email,
		Mail:   m,
		Folder: folder,
		Read:   false,
		Trash:  false,
	}, options.InsertOne().SetBypassDocumentValidation(true))
	return nil
}

func (s *Storage) MoveMail(user string, mailId string, folder string) error {
	mailsCollection := s.db.Collection(qualifiedMailCollection(user))

	oId, err := primitive.ObjectIDFromHex(mailId)
	if err != nil {
		return err
	}

	if folder == common.Trash {
		_, err = mailsCollection.UpdateOne(context.Background(), bson.M{"_id": oId}, bson.M{"$set": bson.M{"trash": true}})
	} else {
		_, err = mailsCollection.UpdateOne(context.Background(), bson.M{"_id": oId}, bson.M{"$set": bson.M{"folder": folder, "trash": false}})
	}
	return err
}

func (s *Storage) RestoreMail(user string, mailId string) error {
	mailsCollection := s.db.Collection(qualifiedMailCollection(user))

	oId, err := primitive.ObjectIDFromHex(mailId)
	if err != nil {
		return err
	}

	//TODO: Legacy for old databases remove soon
	metadata, err := s.GetMail(user, mailId)
	if metadata.Folder == common.Trash {
		_, err = mailsCollection.UpdateOne(context.Background(), bson.M{"_id": oId}, bson.M{"$set": bson.M{"folder": common.Inbox}})
	}

	_, err = mailsCollection.UpdateOne(context.Background(), bson.M{"_id": oId}, bson.M{"$set": bson.M{"trash": false}})
	return err
}

func (s *Storage) DeleteMail(user string, mailId string) error {
	mailsCollection := s.db.Collection(qualifiedMailCollection(user))

	oId, err := primitive.ObjectIDFromHex(mailId)
	if err != nil {
		return err
	}

	_, err = mailsCollection.DeleteOne(context.Background(), bson.M{"_id": oId})
	return err
}

func (s *Storage) GetMailList(user, email, folder string, frame common.Frame) ([]*common.MailMetadata, error) {
	mailsCollection := s.db.Collection(qualifiedMailCollection(user))

	matchFilter := bson.M{"email": email}
	if folder == common.Trash {
		matchFilter["$or"] = bson.A{
			bson.M{"trash": true},
			bson.M{"folder": folder}, //TODO: Legacy for old databases remove soon
		}
	} else {
		matchFilter["folder"] = folder
		matchFilter["$or"] = bson.A{
			bson.M{"trash": false},
			bson.M{"trash": bson.M{"$exists": false}}, //TODO: Legacy for old databases remove soon
		}
	}

	request := bson.A{
		bson.M{"$match": matchFilter},
		bson.M{"$sort": bson.M{"mail.header.date": -1}},
	}

	if frame.Skip > 0 {
		request = append(request, bson.M{"$skip": frame.Skip})
	}

	fmt.Printf("Trying limit number of mails: %v\n", frame)
	if frame.Limit > 0 {
		fmt.Printf("Limit number of mails: %v\n", frame)
		request = append(request, bson.M{"$limit": frame.Limit})
	}

	cur, err := mailsCollection.Aggregate(context.Background(), request)

	if err != nil {
		log.Println(err.Error())
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
		// fmt.Printf("Add mail: %s", result.Id)
		headers = append(headers, result)
	}

	// fmt.Printf("Mails read from database: %v", headers)
	return headers, nil
}

func (s *Storage) GetUserInfo(user string) (*common.UserInfo, error) {
	result := &common.UserInfo{}
	err := s.usersCollection.FindOne(context.Background(), bson.M{"user": user}).Decode(result)
	return result, err
}

func (s *Storage) GetEmailStats(user string, email string, folder string) (unread, total int, err error) {
	mailsCollection := s.db.Collection(qualifiedMailCollection(user))
	result := &struct {
		Total  int
		Unread int
	}{}

	matchFilter := bson.M{"email": email}
	if folder == common.Trash {
		matchFilter["$or"] = bson.A{
			bson.M{"trash": true},
			bson.M{"folder": folder}, //TODO: Legacy for old databases remove soon
		}
	} else {
		matchFilter["folder"] = folder
		matchFilter["$or"] = bson.A{
			bson.M{"trash": false},
			bson.M{"trash": bson.M{"$exists": false}}, //TODO: Legacy for old databases remove soon
		}
	}

	unreadMatchFilter := matchFilter
	unreadMatchFilter["read"] = false

	cur, err := mailsCollection.Aggregate(context.Background(), bson.A{bson.M{"$match": unreadMatchFilter}, bson.M{"$count": "unread"}})
	if err == nil && cur.Next(context.Background()) {
		cur.Decode(result)
	} else {
		return 0, 0, err
	}

	cur, err = mailsCollection.Aggregate(context.Background(), bson.A{bson.M{"$match": matchFilter}, bson.M{"$count": "total"}})
	if err == nil && cur.Next(context.Background()) {
		cur.Decode(result)
	} else {
		return 0, 0, err
	}

	return result.Unread, result.Total, err
}

func (s *Storage) GetMail(user string, id string) (metadata *common.MailMetadata, err error) {
	mailsCollection := s.db.Collection(qualifiedMailCollection(user))

	oId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	metadata = &common.MailMetadata{
		Mail: &common.Mail{},
	}

	err = mailsCollection.FindOne(context.Background(), bson.M{"_id": oId}).Decode(metadata)
	if err != nil {
		return nil, err
	}
	return metadata, nil
}

func (s *Storage) SetRead(user string, id string, read bool) error {
	mailsCollection := s.db.Collection(qualifiedMailCollection(user))

	oId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}
	_, err = mailsCollection.UpdateOne(context.Background(), bson.M{"_id": oId}, bson.M{"$set": bson.M{"read": read}})
	return err
}

func (s *Storage) GetAttachment(user string, attachmentId string) (filePath string, err error) {
	return "", nil
}

func (s *Storage) GetUsers() (users []string, err error) {
	return nil, nil
}

func (s *Storage) GetEmails(user string) (emails []string, err error) {
	result := &struct {
		Email []string
	}{}
	err = s.emailsCollection.FindOne(context.Background(), bson.M{"user": user}).Decode(result)
	if err != nil {
		return nil, err
	}
	return result.Email, nil
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

func (s *Storage) GetFolders(email string) (folders []*common.Folder) {
	folders = []*common.Folder{
		&common.Folder{Name: common.Inbox, Custom: false},
		&common.Folder{Name: common.Trash, Custom: false},
		&common.Folder{Name: common.Spam, Custom: false},
	}
	return
}
