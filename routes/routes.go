package routes

import (
	"iibb-imports/controllers"

	"github.com/gin-gonic/gin"
	"iibb-imports/middleware"
)

func CreateRoutes(r *gin.Engine) {

	v1 := r.Group("/payment/api/v1")
	v1.GET("/healthcheck", controllers.Healthcheck)
	v1.POST("/notificaction-mol-payment", middleware.TokenAuthMiddleware(), controllers.PostPayment)

}
