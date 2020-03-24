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
	"strings"
	"time"

	common "git.semlanik.org/semlanik/gostfix/common"
	"git.semlanik.org/semlanik/gostfix/utils"
	"github.com/semlanik/berkeleydb"
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

	return nil
}

func (s *Storage) UpdateUser(user, password, fullName string) error {
	userInfo := bson.M{}

	if len(password) > 0 && len(password) < 128 {
		hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		hashString := string(hash)
		userInfo["password"] = hashString
	}

	if len(fullName) > 0 && len(fullName) < 128 && utils.RegExpUtilsInstance().FullNameChecker.MatchString(fullName) {
		userInfo["fullName"] = fullName
	}

	if len(userInfo) > 0 {
		_, err := s.usersCollection.UpdateOne(context.Background(), bson.M{"user": user}, bson.M{"$set": userInfo})
		if err != nil {
			return err
		}
	}

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

	emailParts := strings.Split(email, "@")

	if len(emailParts) != 2 {
		return errors.New("Invalid email format")
	}

	db, err := berkeleydb.NewDB()
	if err != nil {
		log.Fatal(err)
	}

	err = db.Open(config.ConfigInstance().VMailboxMaps, berkeleydb.DbHash, 0)
	if err != nil {
		log.Fatalf("Unable to open virtual mailbox maps %s %s\n", config.ConfigInstance().VMailboxMaps, err)
	}
	defer db.Close()

	err = db.Put(email, emailParts[1]+"/"+emailParts[0])
	if err != nil {
		return errors.New("Unable to add email to maps" + err.Error())
	}

	_, err = s.emailsCollection.UpdateOne(context.Background(),
		bson.M{"user": user},
		bson.M{"$addToSet": bson.M{"email": email}},
		options.Update().SetUpsert(upsert))

	return err
}

func (s *Storage) RemoveEmail(user string, email string) error {

	db, err := berkeleydb.NewDB()
	if err != nil {
		log.Fatal(err)
	}

	err = db.Open(config.ConfigInstance().VMailboxMaps, berkeleydb.DbHash, 0)
	if err != nil {
		log.Fatalf("Unable to open virtual mailbox maps %s %s\n", config.ConfigInstance().VMailboxMaps, err)
	}
	defer db.Close()

	err = db.Delete(email)
	if err != nil {
		return errors.New("Unable to remove email from maps" + err.Error())
	}

	_, err = s.emailsCollection.UpdateOne(context.Background(),
		bson.M{"user": user},
		bson.M{"$pull": bson.M{"email": email}})
	return err
}

func (s *Storage) SaveMail(email, folder string, m *common.Mail, read bool) error {
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
		Read:   read,
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
		Mail: common.NewMail(),
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

func (s *Storage) CheckEmailExists(email string) bool {
	result := s.allEmailsCollection.FindOne(context.Background(), bson.M{"emails": email})
	return result.Err() == nil
}

func (s *Storage) GetFolders(email string) (folders []*common.Folder) {
	folders = []*common.Folder{
		&common.Folder{Name: common.Inbox, Custom: false},
		&common.Folder{Name: common.Sent, Custom: false},
		&common.Folder{Name: common.Trash, Custom: false},
		&common.Folder{Name: common.Spam, Custom: false},
	}
	return
}

func (s *Storage) ReadEmailMaps() (map[string]string, error) {
	registredEmails, err := s.GetAllEmails()
	if err != nil {
		return nil, err
	}

	mailPath := config.ConfigInstance().VMailboxBase

	mapsFile := config.ConfigInstance().VMailboxMaps
	if !utils.FileExists(mapsFile) {
		return nil, errors.New("Could not read virtual mailbox maps")
	}

	db, err := berkeleydb.NewDB()
	if err != nil {
		log.Fatal(err)
	}

	err = db.Open(config.ConfigInstance().VMailboxMaps, berkeleydb.DbHash, berkeleydb.DbRdOnly)
	if err != nil {
		return nil, errors.New("Unable to open virtual mailbox maps " + mapsFile + " " + err.Error())
	}
	defer db.Close()

	cursor, err := db.Cursor()
	if err != nil {
		return nil, errors.New("Unable to read virtual mailbox maps " + mapsFile + " " + err.Error())
	}

	emailMaps := make(map[string]string)

	for true {
		email, path, dberr := cursor.GetNext()
		if dberr != nil {
			break
		}
		found := false
		for _, registredEmail := range registredEmails {
			if email == registredEmail {
				found = true
			}
		}
		if !found {
			return nil, errors.New("Found non-registred mailbox <" + email + "> in mail maps. Database has inconsistancy")
		}
		emailMaps[email] = mailPath + "/" + path
	}

	for _, registredEmail := range registredEmails {
		if _, exists := emailMaps[registredEmail]; !exists {
			return nil, errors.New("Found existing mailbox <" + registredEmail + "> in database. Mail maps has inconsistancy")
		}
	}

	return emailMaps, nil
}
