package main

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"

	"github.com/gin-gonic/gin"
)

// downloadImage downloads an image from a URL and saves it to a specified folder.
func downloadImage(imageURL, downloadFolder string) (string, error) {
	// Parse the URL to extract the file name
	parsedURL, err := url.Parse(imageURL)
	if err != nil {
		return "", fmt.Errorf("failed to parse URL %s: %v", imageURL, err)
	}
	fileName := path.Base(parsedURL.Path)
	decodedFileName, err := url.QueryUnescape(fileName)
	if err != nil {
		return "", fmt.Errorf("failed to unescape file name %s: %v", fileName, err)
	}

	// Handle Dropbox shared link
	if strings.Contains(imageURL, "dropbox.com") && strings.Contains(imageURL, "dl=0") {
		imageURL = strings.Replace(imageURL, "dl=0", "raw=1", 1)
	}

	// Create the download folder if it doesn't exist
	if _, err := os.Stat(downloadFolder); os.IsNotExist(err) {
		err := os.MkdirAll(downloadFolder, os.ModePerm)
		if err != nil {
			return "", fmt.Errorf("failed to create download folder %s: %v", downloadFolder, err)
		}
	}

	// Download the image
	resp, err := http.Get(imageURL)
	if err != nil {
		return "", fmt.Errorf("failed to download image from %s: %v", imageURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to download image from %s: %s", imageURL, resp.Status)
	}

	// Save the image to the specified folder
	localPath := path.Join(downloadFolder, decodedFileName)
	file, err := os.Create(localPath)
	if err != nil {
		return "", fmt.Errorf("failed to create file %s: %v", localPath, err)
	}
	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to save image to %s: %v", localPath, err)
	}

	return localPath, nil
}

func main() {
	r := gin.Default()
	r.LoadHTMLGlob("templates/*")

	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", nil)
	})

	r.POST("/download", func(c *gin.Context) {
		urls := c.PostForm("urls")
		imageURLs := strings.Split(urls, "\n")
		downloadFolder := "downloaded_images"

		var downloadedFiles []string
		for _, imageURL := range imageURLs {
			imageURL = strings.TrimSpace(imageURL) // Trim any extra whitespace
			if imageURL == "" {
				continue
			}
			filePath, err := downloadImage(imageURL, downloadFolder)
			if err != nil {
				c.String(http.StatusInternalServerError, fmt.Sprintf("Failed to download image: %v", err))
				return
			}
			downloadedFiles = append(downloadedFiles, filePath)
		}

		c.HTML(http.StatusOK, "result.html", gin.H{
			"files": downloadedFiles,
		})
	})

	r.Static("/downloaded_images", "./downloaded_images") // Serve the downloaded images statically

	r.Run(":8080") // Run on port 8080
}
