package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type Cfg struct {
	Endpoint        string            `json:"endpoint"`
	AccessKeyID     string            `json:"access_key_id"`
	SecretAccessKey string            `json:"secret_access_key"`
	BucketName      string            `json:"bucket_name"`
	IsSSL           bool              `json:"is_ssl"`
	SslCertFile     string            `json:"ssl_cert_file"`
	ContentTypes    map[string]string `json:"content_types"`
}

func Create(endpoint string, accessKeyID string, secretAccessKey string, isSSL bool) (*minio.Client, error) {
	// Initialize minio client object.
	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: isSSL,
	})
	if err != nil {
		return nil, err
	}
	return minioClient, nil
}

func Upload(client *minio.Client, srcFilePath string, bucketName string, contentTypes map[string]string) (string, error) {
	curr := time.Now().Format("20060102150405")
	newFileName := fmt.Sprintf("%s-%s", curr, filepath.Base(srcFilePath))
	extName := filepath.Ext(srcFilePath)
	contentType, exist := contentTypes[extName]
	if !exist {
		return "", errors.New("not support file type")
	}

	_, err := client.FPutObject(context.Background(),
		bucketName,
		newFileName,
		srcFilePath,
		minio.PutObjectOptions{
			ContentType: contentType,
		})
	if err != nil {
		return "", err
	}
	return client.EndpointURL().JoinPath(bucketName, newFileName).String(), nil
}

func downloadFile(sslCertFile string) string {
	resp, err := http.Get(sslCertFile)
	if err != nil {
		fmt.Println(err)
		return ""
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	pid := syscall.Getpid()
	sslCertFile = fmt.Sprintf("/tmp/light_minio_client_%d.cert", pid)
	file, err := os.OpenFile(sslCertFile, os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		fmt.Println("open file failed, err:", err)
		return ""
	}
	defer file.Close()
	file.Write(body)

	return sslCertFile
}

func main() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Printf("%v", err)
		log.Fatalln(err)
		return
	}
	content, err := os.ReadFile(filepath.Join(homeDir, ".light-minio-client.json"))
	if err != nil {
		fmt.Printf("%v", err)
		log.Fatalln(err)
		return
	}
	var cfg Cfg
	err = json.Unmarshal(content, &cfg)
	if err != nil {
		fmt.Printf("%v", err)
		log.Fatalln(err)
		return
	}

	sslCertFile := cfg.SslCertFile
	if strings.HasPrefix(sslCertFile, "http://") ||
		strings.HasPrefix(sslCertFile, "https://") {
		sslCertFile = downloadFile(sslCertFile)
	}
	if len(sslCertFile) > 0 {
		os.Setenv("SSL_CERT_FILE", sslCertFile)
	}

	client, err := Create(cfg.Endpoint, cfg.AccessKeyID, cfg.SecretAccessKey, cfg.IsSSL)
	if err != nil {
		fmt.Printf("%v", err)
		log.Fatalln(err)
		return
	}

	result := ""
	for _, filePath := range os.Args[1:] {
		url, err := Upload(client, filePath, cfg.BucketName, cfg.ContentTypes)
		if err != nil {
			log.Print(err)
			continue
		}
		if result != "" {
			result += "\n"
		}
		result += url
	}
	fmt.Print(result + "\n")
	os.Exit(0)
}
