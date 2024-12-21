package controllers

import (
	"encoding/json"
	"net/http"
	"os"
	"strconv"
	"sync"

	"BSM/models"

	"github.com/gin-gonic/gin"
)

var mu sync.Mutex // Mutex for thread-safe file operations

const jsonFile = "links.json"

// Read data from JSON file
func readJSONFile() ([]models.Link, error) {
	mu.Lock()
	defer mu.Unlock()

	var links []models.Link
	file, err := os.Open(jsonFile)
	if err != nil {
		if os.IsNotExist(err) {
			return links, nil // Return empty slice if file doesn't exist
		}
		return nil, err
	}
	defer file.Close()

	if err := json.NewDecoder(file).Decode(&links); err != nil {
		return nil, err
	}

	return links, nil
}

// Write data to JSON file
func writeJSONFile(links []models.Link) error {
	mu.Lock()
	defer mu.Unlock()

	file, err := os.Create(jsonFile)
	if err != nil {
		return err
	}
	defer file.Close()

	return json.NewEncoder(file).Encode(links)
}

// Get all links
func GetLinks(c *gin.Context) {
	links, err := readJSONFile()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, links)
}

// Create a new link
func CreateLink(c *gin.Context) {
	var link models.Link
	if err := c.ShouldBindJSON(&link); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	links, err := readJSONFile()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	link.ID = uint(len(links) + 1) // Generate a new ID
	links = append(links, link)

	if err := writeJSONFile(links); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, link)
}

// Update a link
func UpdateLink(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	var updatedLink models.Link
	if err := c.ShouldBindJSON(&updatedLink); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	links, err := readJSONFile()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	for i, l := range links {
		if l.ID == uint(id) {
			updatedLink.ID = l.ID // Retain the original ID
			links[i] = updatedLink
			if err := writeJSONFile(links); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, updatedLink)
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{"error": "Link not found"})
}

// Delete a link
func DeleteLink(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	links, err := readJSONFile()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	for i, l := range links {
		if l.ID == uint(id) {
			links = append(links[:i], links[i+1:]...)
			if err := writeJSONFile(links); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, gin.H{"message": "Link deleted successfully"})
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{"error": "Link not found"})
}
