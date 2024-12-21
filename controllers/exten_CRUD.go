package controllers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gocolly/colly/v2"
)

const extensionsJSONFile = "extensions.json"

var extensionsMu sync.Mutex // Mutex for thread-safe file operations

// Extension struct
type Extension struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

// Read extensions from JSON file
func readExtensionsJSONFile() ([]Extension, error) {
	extensionsMu.Lock()
	defer extensionsMu.Unlock()

	var extensions []Extension
	file, err := os.Open(extensionsJSONFile)
	if err != nil {
		if os.IsNotExist(err) {
			return extensions, nil // Return empty slice if file doesn't exist
		}
		return nil, err
	}
	defer file.Close()

	if err := json.NewDecoder(file).Decode(&extensions); err != nil {
		return nil, err
	}

	return extensions, nil
}

// Write extensions to JSON file
func writeExtensionsJSONFile(extensions []Extension) error {
	extensionsMu.Lock()
	defer extensionsMu.Unlock()

	file, err := os.Create(extensionsJSONFile)
	if err != nil {
		return err
	}
	defer file.Close()

	return json.NewEncoder(file).Encode(extensions)
}

// SaveExtension saves an extension name to the JSON file
func SaveExtension(c *gin.Context) {
	var extension Extension
	if err := c.ShouldBindJSON(&extension); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	extensions, err := readExtensionsJSONFile()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	extensions = append(extensions, extension)

	if err := writeExtensionsJSONFile(extensions); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Extension saved successfully", "extension": extension})
}

// GetAllExtensions retrieves all saved extensions from the JSON file
func GetAllExtensions(c *gin.Context) {
	extensions, err := readExtensionsJSONFile()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, extensions)
}

// SearchExtension searches for an extension online and returns download links
func SearchExtension(c *gin.Context) {
	name := c.Query("name")

	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Extension name is required"})
		return
	}

	results := searchAllBrowsers(name)

	// Log the results for debugging
	log.Printf("Search results for '%s': %+v", name, results)

	// Optionally, save the extension name in the JSON file for reference
	extensions, err := readExtensionsJSONFile()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Save each result to the JSON file
	for _, data := range results {
		if data["name"] != "" {
			extension := Extension{Name: data["name"], URL: data["url"]}
			extensions = append(extensions, extension)
		}
	}

	if err := writeExtensionsJSONFile(extensions); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, results)
}

func normalizeName(name string) string {
	return strings.ToLower(strings.ReplaceAll(name, " ", ""))
}

func searchAllBrowsers(name string) map[string]map[string]string {
	normalizedName := normalizeName(name)
	results := make(map[string]map[string]string)

	results["chrome"] = searchChromeWebStore(normalizedName)
	results["firefox"] = searchFirefoxAddons(normalizedName)
	results["safari"] = searchSafariExtensions(normalizedName)

	return results
}

func searchChromeWebStore(name string) map[string]string {
	result := make(map[string]string)
	c := colly.NewCollector()

	c.OnHTML(".h-Ja-d-Ac", func(e *colly.HTMLElement) {
		if len(result) == 0 {
			extractedName := normalizeName(e.ChildText(".h-Ja-d-b"))
			if strings.Contains(extractedName, name) {
				result["name"] = e.ChildText(".h-Ja-d-b")
				result["url"] = fmt.Sprintf("https://chrome.google.com%s", e.ChildAttr("a", "href"))
			}
		}
	})

	searchURL := fmt.Sprintf("https://chrome.google.com/webstore/search/%s?hl=en", url.QueryEscape(name))
	err := c.Visit(searchURL)
	if err != nil {
		log.Printf("Error searching Chrome Web Store: %v", err)
	}

	return result
}

func searchFirefoxAddons(name string) map[string]string {
	result := make(map[string]string)
	c := colly.NewCollector()

	c.OnHTML(".SearchResult", func(e *colly.HTMLElement) {
		if len(result) == 0 {
			result["name"] = strings.TrimSpace(e.ChildText(".SearchResult-name"))
			result["url"] = fmt.Sprintf("https://addons.mozilla.org%s", e.ChildAttr("a.SearchResult-link", "href"))
		}
	})

	searchURL := fmt.Sprintf("https://addons.mozilla.org/en-US/firefox/search/?q=%s", url.QueryEscape(name))
	err := c.Visit(searchURL)
	if err != nil {
		log.Printf("Error searching Firefox Add-ons: %v", err)
	}

	return result
}

func searchSafariExtensions(name string) map[string]string {
	result := make(map[string]string)
	c := colly.NewCollector()

	c.OnHTML(".we-lockup__title", func(e *colly.HTMLElement) {
		if len(result) == 0 {
			extractedName := normalizeName(e.Text)
			if strings.Contains(extractedName, name) {
				result["name"] = e.Text
				result["url"] = fmt.Sprintf("https://apps.apple.com%s", e.ChildAttr("a", "href"))
			}
		}
	})

	searchURL := fmt.Sprintf("https://apps.apple.com/us/search?term=%s%%20safari%%20extension", url.QueryEscape(name))
	err := c.Visit(searchURL)
	if err != nil {
		log.Printf("Error searching Safari Extensions: %v", err)
	}

	return result
}
