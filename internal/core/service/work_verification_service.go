package service

import (
	"context"
	"fmt"
	"log"
	"os"
	"workVerification/internal/adapters/adb"
	"workVerification/internal/adapters/s3"
)

type Service struct {
	fileStorage s3.FileStorage
	adbClient   *adb.ADB
}

func NewService(fs s3.FileStorage, adbClient *adb.ADB) *Service {
	return &Service{
		fileStorage: fs,
		adbClient:   adbClient,
	}
}

func (s *Service) Verify(fileName string) error {
	fmt.Println("Verifying file:", fileName)

	file, err := s.fileStorage.GetFile(fileName)
	if err != nil {
		return err
	}

	ctx := context.Background()

	err = s.adbClient.Verify(ctx, file.Name())
	if err != nil {
		return err
	}

	defer func() {
		err = file.Close()
		if err != nil {
			log.Fatal("failed to close file: ", err.Error())
		}

		err = os.Remove(fileName)
		if err != nil {
			log.Fatal("failed to remove file: ", err.Error())
		}
	}()

	return nil
}
