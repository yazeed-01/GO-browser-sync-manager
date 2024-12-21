package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"BSM/controllers"
)

func Home(c *gin.Context) {
	c.HTML(http.StatusOK, "index.html", gin.H{})
}

func ExtensionPage(c *gin.Context) {
	c.HTML(http.StatusOK, "extension.html", gin.H{})
}

func SetupRoutes() *gin.Engine {
	r := gin.Default()

	// Serve static files
	//r.Static("/static", "./static")
	r.LoadHTMLGlob("templates/*")

	// Home route
	r.GET("/", Home) // Home route

	// Extension page route
	r.GET("/exten", ExtensionPage)

	// Link routes
	r.GET("/links", controllers.GetLinks)          // Get all links
	r.POST("/links", controllers.CreateLink)       // Create a link
	r.PUT("/links/:id", controllers.UpdateLink)    // Update a link
	r.DELETE("/links/:id", controllers.DeleteLink) // Delete a link

	// Email routes
	r.GET("/emails", controllers.GetEmails)          // Get all emails
	r.POST("/emails", controllers.CreateEmail)       // Create an email
	r.PUT("/emails/:id", controllers.UpdateEmail)    // Update an email
	r.DELETE("/emails/:id", controllers.DeleteEmail) // Delete an email

	// Extension routes
	r.POST("/save", controllers.SaveExtension)         // Save extension name to database
	r.GET("/search", controllers.SearchExtension)      // Search extension online and return download link
	r.GET("/extensions", controllers.GetAllExtensions) // Optional: Retrieve all saved extensions

	return r
}
