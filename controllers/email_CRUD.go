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

var emailMu sync.Mutex // Mutex for thread-safe file operations
const emailJSONFile = "emails.json"

// Read emails from JSON file
func readEmailJSONFile() ([]models.Email, error) {
	emailMu.Lock()
	defer emailMu.Unlock()

	var emails []models.Email
	file, err := os.Open(emailJSONFile)
	if err != nil {
		if os.IsNotExist(err) {
			return emails, nil // Return empty slice if file doesn't exist
		}
		return nil, err
	}
	defer file.Close()

	if err := json.NewDecoder(file).Decode(&emails); err != nil {
		return nil, err
	}

	return emails, nil
}

// Write emails to JSON file
func writeEmailJSONFile(emails []models.Email) error {
	emailMu.Lock()
	defer emailMu.Unlock()

	file, err := os.Create(emailJSONFile)
	if err != nil {
		return err
	}
	defer file.Close()

	return json.NewEncoder(file).Encode(emails)
}

// Get all emails
func GetEmails(c *gin.Context) {
	emails, err := readEmailJSONFile()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, emails)
}

// Create a new email
func CreateEmail(c *gin.Context) {
	var email models.Email
	if err := c.ShouldBindJSON(&email); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	emails, err := readEmailJSONFile()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	email.ID = uint(len(emails) + 1) // Generate a new ID
	emails = append(emails, email)

	if err := writeEmailJSONFile(emails); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, email)
}

// Update an email
func UpdateEmail(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	var email models.Email
	if err := c.ShouldBindJSON(&email); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	emails, err := readEmailJSONFile()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	for i, e := range emails {
		if e.ID == uint(id) {
			emails[i] = email
			if err := writeEmailJSONFile(emails); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, email)
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{"error": "Email not found"})
}

// Delete an email
func DeleteEmail(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	emails, err := readEmailJSONFile()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	for i, e := range emails {
		if e.ID == uint(id) {
			emails = append(emails[:i], emails[i+1:]...)
			if err := writeEmailJSONFile(emails); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, gin.H{"message": "Email deleted successfully"})
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{"error": "Email not found"})
}
