package efeed

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"path/filepath"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
)

func downloadFile(filepath string, url string) error {
	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		fmt.Println("downloading: " + filepath)
		// Get the data
		resp, err := http.Get(url)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		// Create the file
		out, err := os.Create(filepath)
		if err != nil {
			return err
		}
		defer out.Close()

		// Write the body to file
		_, err = io.Copy(out, resp.Body)
		return err
	}
	return nil
}

func deleteFile(path string) {
	_ = os.Remove(path)
}

// UploadToDO UploadToDO
func UploadToDO(spaceURL, siteFolder, link string, svc *s3.S3) (string, error) {
	fmt.Println("uploading image: ", link)
	imagesFolder := filepath.Join(".", "images")
	os.MkdirAll(imagesFolder, os.ModePerm)
	imagesSiteFolder := filepath.Join(imagesFolder, siteFolder)
	os.MkdirAll(imagesSiteFolder, os.ModePerm)

	var fileName string
	switch siteFolder {
	case "fanatics":
		fileName = getFanaticsFileName(link)
	case "soccerpro":
		fileName = getSoccerProFileName(link)
	}
	imagesPath := filepath.Join(imagesSiteFolder, fileName)
	downloadFile(imagesPath, link)
	defer deleteFile(imagesPath)
	var imageURL string

	uploadToDO(siteFolder, fileName, imagesPath, "efeed", svc)
	imageURL = spaceURL + siteFolder + "/" + fileName
	return imageURL, nil
}

func getFanaticsFileName(link string) string {
	linkWithoutParamsArr := strings.Split(link, "&")
	linkWithoutParams := linkWithoutParamsArr[len(linkWithoutParamsArr)-2]
	linkParts := strings.Split(linkWithoutParams, "/")
	name := linkParts[len(linkParts)-1]
	return name
}

func getSoccerProFileName(link string) string {
	linkParts := strings.Split(link, "/")
	name := linkParts[len(linkParts)-1]
	return name
}

func uploadToDO(siteFolder, fileName, path string, bucket string, svc *s3.S3) {
	fmt.Println("uploading: ", path)
	file, _ := os.Open(path)
	defer file.Close()
	fileInfo, _ := file.Stat()
	size := fileInfo.Size()
	buffer := make([]byte, size) // read file content to buffer

	file.Read(buffer)
	fileBytes := bytes.NewReader(buffer)
	fileType := http.DetectContentType(buffer)

	params := &s3.PutObjectInput{
		Bucket:        aws.String(bucket),
		Key:           aws.String(siteFolder + "/" + fileName),
		Body:          fileBytes,
		ContentLength: aws.Int64(size),
		ContentType:   aws.String(fileType),
		ACL:           aws.String("public-read"),
	}
	_, _ = svc.PutObject(params)
}
