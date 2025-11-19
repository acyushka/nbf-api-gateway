package s3

import (
	models "api-gateway/internal/ports/handlers/user_handler"
	"context"
	"io"

	s3v1 "github.com/acyushka/nbf-file-storage-service/pkg/pb/gen"

	authInt "github.com/hesoyamTM/nbf-auth/pkg/auth"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type FileStorageClient struct {
	api s3v1.FileStorageServiceClient
}

func New(ctx context.Context, address string) (*FileStorageClient, error) {
	cc, err := grpc.NewClient(address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(authInt.SettingMetadataInterceptor()),
	)
	if err != nil {
		return nil, err
	}

	return &FileStorageClient{
		api: s3v1.NewFileStorageServiceClient(cc),
	}, nil
}

func (c *FileStorageClient) UploadAvatar(ctx context.Context, userID string, file *models.FilePhoto) (string, error) {
	fileData, err := io.ReadAll(file.Data)
	if err != nil {
		return "", err
	}

	resp, err := c.api.UploadAvatar(ctx, &s3v1.UploadAvatarRequest{
		UserId:      userID,
		FileData:    fileData,
		FileName:    file.FileName,
		ContentType: file.ContentType,
	})
	if err != nil {
		return "", err
	}

	return resp.GetPhotoId(), nil
}

func (c *FileStorageClient) UploadPhotos(ctx context.Context, userID string, files []*models.FilePhoto) ([]string, error) {
	photos := make([]*s3v1.Photo, 0, len(files))

	for _, file := range files {
		fileData, err := io.ReadAll(file.Data)
		if err != nil {
			return nil, err
		}

		photos = append(photos, &s3v1.Photo{
			FileData:    fileData,
			FileName:    file.FileName,
			ContentType: file.ContentType,
		})
	}

	resp, err := c.api.UploadPhotos(ctx, &s3v1.UploadPhotosRequest{
		UserId: userID,
		Photos: photos,
	})
	if err != nil {
		return nil, err
	}

	return resp.GetPhotoIds(), nil
}

func (c *FileStorageClient) GetPhotoURL(ctx context.Context, userID string, photoID string) (string, error) {
	resp, err := c.api.GetPhotoURL(ctx, &s3v1.GetPhotoURLRequest{
		UserId:  userID,
		PhotoId: photoID,
	})
	if err != nil {
		return "", err
	}

	return resp.GetUrl(), nil
}
