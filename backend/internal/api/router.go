package api

import (
	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()

	// Define routes
	r.GET("/", welcomeHandler)
	r.GET("/fetch-profiles", fetchProfilesHandler)
	r.GET("/update-profiles", updateProfilesHandler)
	r.POST("/add-profile", addProfileHandler)
	r.POST("/execute-profile", executeProfileHandler)

	return r
}
