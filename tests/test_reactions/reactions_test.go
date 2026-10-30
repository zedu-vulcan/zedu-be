package test_reactions

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/hngprojects/telex_be/external/request"
	"github.com/hngprojects/telex_be/internal/models"
	"github.com/hngprojects/telex_be/pkg/controller/auth"
	"github.com/hngprojects/telex_be/tests"
	"github.com/hngprojects/telex_be/utility"
)

func TestReactionNonMemberForbidden(t *testing.T) {
	r, reactionsController := SetupReactionsTestRouter()
	db := reactionsController.Db.Postgresql
	currUUID := utility.GenerateUUID()

	authCtrl := auth.Controller{
		Db:        reactionsController.Db,
		Validator: reactionsController.Validator,
		Logger:    reactionsController.Logger,
		ExtReq: request.ExternalRequest{
			Logger: reactionsController.Logger,
			Test:   true,
		},
	}

	memberSignUpData := models.CreateUserRequestModel{
		Email:       fmt.Sprintf("member%v@qa.team", currUUID),
		Password:    "password123",
		UserName:    fmt.Sprintf("member%v", currUUID),
		FirstName:   "Member",
		LastName:    "User",
		PhoneNumber: fmt.Sprintf("0801%v", currUUID[:7]),
	}

	nonMemberSignUpData := models.CreateUserRequestModel{
		Email:       fmt.Sprintf("nonmember%v@qa.team", currUUID),
		Password:    "password123",
		UserName:    fmt.Sprintf("nonmember%v", currUUID),
		FirstName:   "NonMember",
		LastName:    "User",
		PhoneNumber: fmt.Sprintf("0802%v", currUUID[:7]),
	}

	tests.SignupUser(t, gin.Default(), authCtrl, memberSignUpData, false)
	tests.SignupUser(t, gin.Default(), authCtrl, nonMemberSignUpData, false)

	memberLoginData := models.LoginRequestModel{
		Email:    memberSignUpData.Email,
		Password: memberSignUpData.Password,
	}

	nonMemberLoginData := models.LoginRequestModel{
		Email:    nonMemberSignUpData.Email,
		Password: nonMemberSignUpData.Password,
	}

	memberToken := tests.GetLoginToken(t, gin.Default(), authCtrl, memberLoginData)
	nonMemberToken := tests.GetLoginToken(t, gin.Default(), authCtrl, nonMemberLoginData)

	memberID := tests.GetUserIDFromToken(t, memberToken, db)

	// Create channel and assign only member to it
	channel := models.Channels{
		ID:          utility.GenerateUUID(),
		Name:        "private-channel",
		Description: "Private channel test",
		OwnerId:     memberID,
	}
	db.Create(&channel)

	userChan := models.UserChannels{
		UserID:     memberID,
		ChannelsID: channel.ID,
		Username:   memberSignUpData.UserName,
	}
	db.Create(&userChan)

	// Non-member attempts to post a reaction to channel
	reqBody := models.ReactionRequest{
		ThreadID: utility.GenerateUUID(),
		Reaction: "👍",
		Type:     "thread",
	}
	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/reactions/%s", channel.ID), bytes.NewBuffer(body))
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %v", nonMemberToken))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	tests.AssertStatusCode(t, rr.Code, http.StatusForbidden)
}
