// Copyright 2022-2023 The Memphis.dev Authors
// Licensed under the Memphis Business Source License 1.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// Changed License: [Apache License, Version 2.0 (https://www.apache.org/licenses/LICENSE-2.0), as published by the Apache Foundation.
//
// https://github.com/memphisdev/memphis/blob/master/LICENSE
//
// Additional Use Grant: You may make use of the Licensed Work (i) only as part of your own product or service, provided it is not a message broker or a message queue product or service; and (ii) provided that you do not use, provide, distribute, or make available the Licensed Work as a Service.
// A "Service" is a commercial offering, product, hosted, or managed service, that allows third parties (other than your own employees and contractors acting on your behalf) to access and/or use the Licensed Work or a substantial set of the features or functionality of the Licensed Work to third parties as a software-as-a-service, platform-as-a-service, infrastructure-as-a-service or other similar services that compete with Licensor products or services.
package server

import (
	"encoding/json"
	"fmt"
	"math/rand"

	"github.com/memphisdev/memphis/db"
	"github.com/memphisdev/memphis/memphis_cache"
	"github.com/memphisdev/memphis/models"
	"github.com/memphisdev/memphis/utils"
	"golang.org/x/crypto/bcrypt"

	"github.com/gin-gonic/gin"
)

type AccessTokenHandler struct{ S *Server }

const (
	accessKeyIdLen = 15
	accessKeyLen   = 30
)

func generateRandomString(n int) string {
	var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890")

	b := make([]rune, n)

	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func generateAccessKeyID() string {
	return generateRandomString(accessKeyIdLen)
}

func generateSecretKey() string {
	return generateRandomString(accessKeyLen)
}

func generateAccessToken(userName, description, tenantName string) (*createAccessTokenResp, error) {
	accessKeyID := generateAccessKeyID()
	secretKey := generateSecretKey()

	_, user, err := memphis_cache.GetUser(userName, tenantName, false)
	if err != nil {
		return nil, err
	}

	hashedSecretKey, err := bcrypt.GenerateFromPassword([]byte(secretKey), bcrypt.MinCost)
	if err != nil {
		serv.Errorf("[tenant: %v][user: %v]CreateAccessToken at GenerateFromPassword: AccessToken %v: %v", user.TenantName, user.Username, secretKey, err.Error())
		return nil, err
	}

	err = db.InsertNewAccessToken(user.ID, accessKeyID, string(hashedSecretKey), description, tenantName)
	if err != nil {
		serv.Errorf("[tenant: %v][user: %v]CreateAccessToken at db.InsertNewAccessToken: AccessToken %v: %v", user.TenantName, user.Username, accessKeyID, err.Error())
		return nil, err
	}

	message := fmt.Sprintf("[tenant: %v][user: %v]New AccessToken %v has been created ", user.TenantName, user.Username, accessKeyID)
	serv.Noticef(message)

	return &createAccessTokenResp{
		AccessKeyID: accessKeyID,
		SecretKey:   secretKey,
	}, nil
}

func (at AccessTokenHandler) CreateNewAccessToken(c *gin.Context) {
	var body models.CreateAccessTokenSchema
	ok := utils.Validate(c, &body, false, nil)
	if !ok {
		return
	}

	user, err := getUserDetailsFromMiddleware(c)
	tenantName := user.TenantName
	if err != nil {
		serv.Errorf("CreateNewAccessToken at getUserDetailsFromMiddleware: AccessToken: %v", err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	accessTokenData, err := generateAccessToken(user.Username, body.Description, tenantName)
	if err != nil {
		serv.Errorf("[tenant: %v][user: %v]CreateNewAccessToken at generateAccessToken: AccessToken: %v", tenantName, user.Username, err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
	}

	c.IndentedJSON(200, gin.H{
		"access_key_id": accessTokenData.AccessKeyID,
		"secret_key":    accessTokenData.SecretKey,
	})
}

func (at AccessTokenHandler) GetAllAccessTokens(c *gin.Context) {
	user, err := getUserDetailsFromMiddleware(c)
	tenantName := user.TenantName
	if err != nil {
		serv.Errorf("GetAllAccessTokens at getUserDetailsFromMiddleware: AccessToken: %v", err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	// ASK: should we filter by user tenant name?
	accessTokens, err := db.GetAllAccessTokens()
	if err != nil {
		serv.Errorf("[tenant: %v][user: %v]CreateAccessToken at GetAllAccessTokens: %v: %v", tenantName, user.Username, err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	c.IndentedJSON(200, accessTokens)
}

func (s *Server) createAccessTokenDirect(c *client, reply string, msg []byte) {
	var cat createAccessTokenReq
	var resp createAccessTokenResp

	tenantName, message, err := s.getTenantNameAndMessage(msg)
	if err != nil {
		s.Errorf("createAccessTokenDirect at getTenantNameAndMessage: %v", err.Error())
		respondWithErr(s.MemphisGlobalAccountString(), s, reply, err)
		return
	}

	if err := json.Unmarshal([]byte(message), &cat); err != nil {
		s.Errorf("[tenant: %v][user: %v]createAccessTokenDirect at json.Unmarshal: failed to generate new accessToken %v: %v", tenantName, cat.Username, message, err.Error())
		respondWithErr(s.MemphisGlobalAccountString(), s, reply, err)
		return
	}

	tokenResp, err := generateAccessToken(cat.Username, cat.Description, tenantName)
	if err != nil {
		serv.Errorf("[tenant: %v][user: %v]createAccessTokenDirect at generateAccessToken: failed to generate new accessToken: %v", tenantName, cat.Username, err.Error())
		respondWithRespErr(s.MemphisGlobalAccountString(), s, reply, err, &resp)
	}

	respondWithResp(s.MemphisGlobalAccountString(), s, reply, tokenResp)
}

func (s *Server) validateAccessTokenDirect(c *client, reply string, msg []byte) {
	var req validateAccessTokenReq
	var resp validateAccessTokenResp

	if err := json.Unmarshal(msg, &req); err != nil {
		s.Errorf("validateAccessTokenDirect at json.Unmarshal: failed validate access token: %v", err.Error())
		respondWithErr(s.MemphisGlobalAccountString(), s, reply, err)
		return
	}

	found, accessTokenData, err := db.GetAccessTokenByAccessKeyId(req.AccessKeyID)
	if !found || err != nil || !accessTokenData.IsActive {
		serv.Errorf("validateAccessTokenDirect at GetAccessTokenByAccessKeyId: failed validate access token, found: %v, %v", found, err.Error())
		respondWithRespErr(s.MemphisGlobalAccountString(), s, reply, err, &resp)
	}

	hashedSecretKey := []byte(accessTokenData.SecretKey)
	err = bcrypt.CompareHashAndPassword(hashedSecretKey, []byte(req.SecretKey))
	if err != nil {
		serv.Errorf("validateAccessTokenDirect at GetAccessTokenByAccessKeyId: failed to validate secret key %v", err.Error())
		respondWithRespErr(s.MemphisGlobalAccountString(), s, reply, err, &validateAccessTokenResp{IsValid: false})
	}

	respondWithResp(s.MemphisGlobalAccountString(), s, reply, &validateAccessTokenResp{IsValid: true})
}
