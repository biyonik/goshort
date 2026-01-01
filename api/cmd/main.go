package cmd

import (
	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	r.GET("/health", healthCheck)
	r.GET("/:code", redirectURL)

	_ = r.Run(":8080")
}

func healthCheck(c *gin.Context) {
	c.JSON(200, gin.H{"status": "OK", "message": "Çalışıyor kankaaa"})
}

func redirectURL(c *gin.Context) {
	code := c.Param("code")
	_ = code // TODO: Database'den long URL'i bul

	longURL := "https://google.com"
	c.Redirect(302, longURL)
}
