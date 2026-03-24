package service

import (
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"mime/multipart"
	"os"
	"path/filepath"
	"time"

	"golang.org/x/image/draw"

	"shoe-store/internal/model"
	"shoe-store/internal/repository"
)

type ProductService struct {
	Repo       *repository.ProductRepo
	uploadsDir string
}

func NewProductService(repo *repository.ProductRepo, uploadsDir string) *ProductService {
	return &ProductService{Repo: repo, uploadsDir: uploadsDir}
}

func (s *ProductService) List(filter model.ProductFilter) ([]model.Product, error) {
	return s.Repo.List(filter)
}

func (s *ProductService) GetByID(id int64) (*model.Product, error) {
	p, err := s.Repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if p == nil {
		return nil, errors.New("product not found")
	}
	return p, nil
}

func validateProductInput(input model.ProductInput) error {
	if input.Name == "" {
		return errors.New("name is required")
	}
	if input.Price < 0 {
		return errors.New("price must be >= 0")
	}
	if input.Quantity < 0 {
		return errors.New("quantity must be >= 0")
	}
	if input.CategoryID <= 0 {
		return errors.New("categoryId must be > 0")
	}
	if input.ManufacturerID <= 0 {
		return errors.New("manufacturerId must be > 0")
	}
	if input.SupplierID <= 0 {
		return errors.New("supplierId must be > 0")
	}
	if input.UnitID <= 0 {
		return errors.New("unitId must be > 0")
	}
	return nil
}

func (s *ProductService) Create(input model.ProductInput) (int64, error) {
	if err := validateProductInput(input); err != nil {
		return 0, err
	}
	return s.Repo.Create(input)
}

func (s *ProductService) Update(id int64, input model.ProductInput) error {
	if err := validateProductInput(input); err != nil {
		return err
	}
	return s.Repo.Update(id, input)
}

func (s *ProductService) Delete(id int64) error {
	err := s.Repo.Delete(id)
	if err != nil {
		if containsForeignKeyErr(err) {
			return errors.New("Товар присутствует в заказе и не может быть удалён")
		}
		return err
	}
	return nil
}

func containsForeignKeyErr(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return contains(msg, "FOREIGN KEY") || contains(msg, "foreign key")
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsStr(s, substr))
}

func containsStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func (s *ProductService) UploadImage(id int64, file multipart.File, header *multipart.FileHeader) error {
	contentType := header.Header.Get("Content-Type")
	if contentType != "image/jpeg" && contentType != "image/png" {
		return errors.New("only image/jpeg and image/png are allowed")
	}

	var src image.Image
	var decodeErr error
	if contentType == "image/jpeg" {
		src, decodeErr = jpeg.Decode(file)
	} else {
		src, decodeErr = png.Decode(file)
	}
	if decodeErr != nil {
		return fmt.Errorf("failed to decode image: %w", decodeErr)
	}

	// Resize to fit within 300x200 preserving aspect ratio
	dst := resizeToFit(src, 300, 200)

	filename := fmt.Sprintf("product_%d_%d.jpg", id, time.Now().UnixNano())
	outPath := filepath.Join(s.uploadsDir, filename)

	if err := os.MkdirAll(s.uploadsDir, 0755); err != nil {
		return fmt.Errorf("failed to create uploads dir: %w", err)
	}

	outFile, err := os.Create(outPath)
	if err != nil {
		return fmt.Errorf("failed to create image file: %w", err)
	}
	defer outFile.Close()

	if err := jpeg.Encode(outFile, dst, &jpeg.Options{Quality: 85}); err != nil {
		return fmt.Errorf("failed to encode image: %w", err)
	}

	// Delete old image if exists
	oldPath, err := s.Repo.GetImage(id)
	if err == nil && oldPath != "" {
		os.Remove(filepath.Join(s.uploadsDir, filepath.Base(oldPath)))
	}

	return s.Repo.UpdateImage(id, filename)
}

func resizeToFit(src image.Image, maxW, maxH int) image.Image {
	bounds := src.Bounds()
	srcW := bounds.Dx()
	srcH := bounds.Dy()

	if srcW == 0 || srcH == 0 {
		return src
	}

	newW, newH := srcW, srcH
	if newW > maxW {
		newH = newH * maxW / newW
		newW = maxW
	}
	if newH > maxH {
		newW = newW * maxH / newH
		newH = maxH
	}

	if newW == srcW && newH == srcH {
		return src
	}

	dst := image.NewRGBA(image.Rect(0, 0, newW, newH))
	draw.BiLinear.Scale(dst, dst.Bounds(), src, src.Bounds(), draw.Over, nil)
	return dst
}
