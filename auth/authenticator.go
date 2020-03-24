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

package auth

import (
	"context"
	"errors"
	"log"
	"time"

	"git.semlanik.org/semlanik/gostfix/config"
	utils "git.semlanik.org/semlanik/gostfix/utils"
	uuid "github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"
)

type Authenticator struct {
	db               *mongo.Database
	usersCollection  *mongo.Collection
	tokensCollection *mongo.Collection
}

type Privileges int

const (
	AdminPrivilege = 1 << iota
	SendMailPrivilege
)

func NewAuthenticator() (*Authenticator, error) {
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
	a := &Authenticator{
		db:               db,
		usersCollection:  db.Collection("users"),
		tokensCollection: db.Collection("tokens"),
	}
	return a, nil
}

func (a *Authenticator) CheckUser(user, password string) error {
	log.Printf("Check user: %s", user)
	result := struct {
		User     string
		Password string
	}{}
	err := a.usersCollection.FindOne(context.Background(), bson.M{"user": user}).Decode(&result)
	if err != nil {
		return errors.New("Invalid user or password")
	}

	if bcrypt.CompareHashAndPassword([]byte(result.Password), []byte(password)) != nil {
		return errors.New("Invalid user or password")
	}
	return nil
}

func (a *Authenticator) addToken(user, token string) error {
	log.Printf("Add token: %s\n", user)
	a.tokensCollection.UpdateOne(context.Background(),
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
	a.cleanupTokens(user)
	return nil
}

func (a *Authenticator) cleanupTokens(user string) {
	log.Printf("Cleanup tokens: %s\n", user)

	cur, err := a.tokensCollection.Aggregate(context.Background(),
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

	_, err = a.tokensCollection.UpdateOne(context.Background(), bson.M{"user": user}, bson.M{"$set": bson.M{"token": tokensToKeep}})
	return
}

func (a *Authenticator) Login(user, password string) (string, bool) {
	if !utils.RegExpUtilsInstance().EmailChecker.MatchString(user) {
		return "", false
	}

	if a.CheckUser(user, password) != nil {
		return "", false
	}

	token := uuid.New().String()
	a.addToken(user, token)
	return token, true
}

func (a *Authenticator) Logout(user, token string) error {
	a.cleanupTokens(user)

	_, err := a.tokensCollection.UpdateOne(context.Background(), bson.M{"user": user}, bson.M{"$pull": bson.M{"token": bson.M{"token": token}}})
	if err != nil {
		log.Printf("Unable to remove token %s", err)
	}

	return err
}

func (a *Authenticator) checkToken(user, token string) error {
	if token == "" {
		return errors.New("Invalid token")
	}

	cur, err := a.tokensCollection.Aggregate(context.Background(),
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

func (a *Authenticator) Verify(user, token string) bool {
	if !utils.RegExpUtilsInstance().EmailChecker.MatchString(user) {
		return false
	}

	return a.checkToken(user, token) == nil
}

func (a *Authenticator) CheckPrivileges(user string, privilege Privileges) {

}
