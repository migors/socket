package tg

import (
	"errors"
	"fmt"
	"io"
	"os"
)

type GetFileResponse struct {
	FileId   string `json:"file_id"`
	FileSize int64  `json:"file_size"`
	FilePath string `json:"file_path"`
}

func GetFile(fileId string, targetPath string) error {
	var fileInfo GetFileResponse
	err := request("getFile", map[string]string{"file_id": fileId}, &fileInfo)
	if err != nil {
		return err
	}

	downloadUrl := "https://api.telegram.org/file/bot" + token + "/" + fileInfo.FilePath
	res, err := client.Get(downloadUrl)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	file, err := os.Create(targetPath)
	if err != nil {
		return err
	}
	defer file.Close()

	written, err := io.Copy(file, res.Body)
	if err != nil {
		return err
	}
	if written != fileInfo.FileSize {
		return errors.New(fmt.Sprintf("File size mismatch, downloaded %d but should be %d", written, fileInfo.FileSize))
	}

	return nil
}
