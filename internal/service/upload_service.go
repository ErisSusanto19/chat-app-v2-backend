package service

import (
	"context"
	"log"
	"mime/multipart"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
)

type UploadResult struct {
	URL      string `json:"url"`
	PublicID string `json:"public_id"`
}

type UploadService interface {
	UploadImage(ctx context.Context, file multipart.File, fileHeader *multipart.FileHeader) (*UploadResult, error)
	DeleteImage(ctx context.Context, publicID string) error
}

type cloudinaryUploadService struct {
	cld *cloudinary.Cloudinary
}

func NewCloudinaryUploadService(cld *cloudinary.Cloudinary) UploadService {
	return &cloudinaryUploadService{cld: cld}
}

func (s *cloudinaryUploadService) UploadImage(ctx context.Context, file multipart.File, fileHeader *multipart.FileHeader) (*UploadResult, error) {
	uploadParams := uploader.UploadParams{
		Folder:         "chat-app-avatars",
		Transformation: "w_400,h_400,c_fill,g_face/q_auto",
	}

	uploadResult, err := s.cld.Upload.Upload(ctx, file, uploadParams)
	if err != nil {
		return nil, err
	}

	return &UploadResult{
		URL:      uploadResult.SecureURL,
		PublicID: uploadResult.PublicID,
	}, nil
}

func (s *cloudinaryUploadService) DeleteImage(ctx context.Context, publicID string) error {
	_, err := s.cld.Upload.Destroy(ctx, uploader.DestroyParams{PublicID: publicID})
	if err != nil {
		log.Printf("Failed to delete image from Cloudinary. PublicID: %s, Error: %v", publicID, err)
	}
	return err
}
