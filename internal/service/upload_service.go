package service

import (
	"context"
	"mime/multipart"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
)

type UploadService interface {
	UploadImage(ctx context.Context, file multipart.File, fileHeader *multipart.FileHeader) (string, error)
}

type cloudinaryUploadService struct {
	cld *cloudinary.Cloudinary
}

func NewCloudinaryUploadService(cld *cloudinary.Cloudinary) UploadService {
	return &cloudinaryUploadService{cld: cld}
}

func (s *cloudinaryUploadService) UploadImage(ctx context.Context, file multipart.File, fileHeader *multipart.FileHeader) (string, error) {
	uploadParams := uploader.UploadParams{
		Folder:         "chat-app-avatars",
		Transformation: "w_400,h_400,c_fill,g_face/q_auto",
	}

	uploadResult, err := s.cld.Upload.Upload(ctx, file, uploadParams)
	if err != nil {
		return "", err
	}

	return uploadResult.SecureURL, nil
}
