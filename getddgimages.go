package getddgimages

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

type images struct {
	Images []Image `json:"results"`
}

// Image struct hold key value pairs that are fetched from
// the duckduckgo api
type Image struct {
	Source    string `json:"source"`
	Title     string `json:"title"`
	Height    int    `json:"height"`
	Width     int    `json:"width"`
	URL       string `json:"url"`
	Image     string `json:"image"`
	Thumbnail string `json:"thumbnail"`
}

const (
	ddgImagesResulCount int = 100
)

// DownloadImages downloads requested number of images from the duckduckgo
// images and writes them to disk under the `./<query>}` folder, returns a
// list of physical paths to the downloaded images.
func DownloadImages(query string, numberOfImages int) ([]string, error) {
	// Get the current path
	currentDir, osErr := os.Getwd()
	if osErr != nil {
		return nil, osErr
	}

	// Create the target dir
	targetDir := filepath.Join(currentDir, url.QueryEscape(query))
	mkdirErr := os.Mkdir(targetDir, 0744)
	if mkdirErr != nil {
		return nil, mkdirErr
	}

	var downloadedFilePaths []string = make([]string, 0)
	startOffset := 0
	commonImageExtension := []string{".jpg", ".jpeg", ".gif", ".png", ".bmp", ".svg", ".webp", ".ico"}
	// Keep crawling unless the requested image number is met
	for len(downloadedFilePaths) < numberOfImages {
		currentImages, crawlErr := GetImageURLs(query, startOffset)
		if crawlErr != nil {
			return nil, crawlErr
		}
		for _, img := range *currentImages {
			// Fetch the image
			currentImage, downloadErr := getImageFromURL(img.Image)
			if downloadErr != nil {
				continue
			}

			targetFileExtension := ".jpg"
			for _, v := range commonImageExtension {
				if strings.HasSuffix(img.Image, v) {
					targetFileExtension = v
					break
				}
			}
			finalFileName := url.QueryEscape(img.Title + targetFileExtension)
			targetFilePath := filepath.Join(targetDir, finalFileName)

			// Write the image to the disk
			writeErr := writeImageToDisk(*currentImage, targetFilePath)
			if writeErr != nil {
				fmt.Println(writeErr)
				continue
			}

			downloadedFilePaths = append(downloadedFilePaths, targetFilePath)

			if len(downloadedFilePaths) >= numberOfImages {
				break
			}
		}

		startOffset += ddgImagesResulCount
	}
	return downloadedFilePaths, nil
}

// GetImageURLs crawls duckduckgo images and returns 100 image URLs related to the `query`
// starting at the given `start` index. The `start` parameter indicates the first
// image's index at the results page.
// e.g. if `start` is set to 15, the function will ignore first 15 image results
// and fetch images at indexes [15, 115]
func GetImageURLs(query string, start int) (*[]Image, error) {

	urlEncodedQuery := url.QueryEscape(query)
	targetURL := fmt.Sprintf("https://duckduckgo.com/?q=%s&iax=images&ia=images", urlEncodedQuery)

	resp, respErr := http.Get(targetURL)
	if respErr != nil {
		return nil, respErr
	}
	defer resp.Body.Close()

	document, readErr := ioutil.ReadAll(resp.Body)
	if readErr != nil {
		return nil, readErr
	}

	// Get the vqd value which is required for requesting the image urls
	stringDocument := string(document)
	vqd := stringDocument[strings.Index(stringDocument, "vqd='")+5:]
	vqd = vqd[:strings.Index(vqd, "'")]

	// Fetch the image urls
	targetURL = fmt.Sprintf("https://duckduckgo.com/i.js?l=us-en&o=json&q=%s&vqd=%s&f=,,,&s=%d", urlEncodedQuery, vqd, start)
	apiResponse, apiErr := http.Get(targetURL)
	if apiErr != nil {
		return nil, apiErr
	}

	body, bodyReadErr := ioutil.ReadAll(apiResponse.Body)
	if bodyReadErr != nil {
		return nil, bodyReadErr
	}

	// Construct images from json string
	var images images
	json.Unmarshal(body, &images)
	return &images.Images, nil
}

func getImageFromURL(url string) (*[]byte, error) {
	response, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode != 200 {
		return nil, errors.New("Received non 200 response code")
	}
	if response.Body == nil {
		return nil, errors.New("Response body is empty")
	}

	// Read the image as byte array from the response body
	rawImage, readErr := ioutil.ReadAll(response.Body)
	if readErr != nil {
		return nil, readErr
	}

	return &rawImage, nil
}

func writeImageToDisk(image []byte, path string) error {
	return ioutil.WriteFile(path, image, 0744)
}
