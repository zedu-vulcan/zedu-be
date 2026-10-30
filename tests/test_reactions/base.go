package test_reactions

import (
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/hngprojects/telex_be/external/request"
	"github.com/hngprojects/telex_be/pkg/controller/reactions"
	"github.com/hngprojects/telex_be/pkg/middleware"
	"github.com/hngprojects/telex_be/pkg/repository/storage"
	"github.com/hngprojects/telex_be/tests"
	"github.com/hngprojects/telex_be/utility"
)

func SetupReactionsTestRouter() (*gin.Engine, *reactions.Controller) {
	gin.SetMode(gin.TestMode)

	logger := tests.Setup()
	db := storage.Connection()
	validator := validator.New()
	utility.RegisterCustomValidations(validator)

	reactionsController := &reactions.Controller{
		Db:        db,
		Validator: validator,
		Logger:    logger,
		ExtReq: request.ExternalRequest{
			Logger: logger,
			Test:   true,
		},
	}

	r := gin.Default()
	r.POST("/api/v1/reactions/:channel_id", middleware.Authorize(reactionsController.Db.Postgresql), reactionsController.CreateReaction)
	return r, reactionsController
}
