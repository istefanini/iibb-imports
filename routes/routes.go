package routes

import (
	"github.com/istefanini/iibb-import/controllers"

	"github.com/gin-gonic/gin"
	"github.com/istefanini/iibb-import/middleware"
)

func CreateRoutes(r *gin.Engine) {

	v1 := r.Group("/payment/api/v1")
	v1.GET("/healthcheck", controllers.Healthcheck)
	v1.POST("/notificaction-mol-payment", middleware.TokenAuthMiddleware(), controllers.PostPayment)

}
