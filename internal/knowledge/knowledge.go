package knowledge

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"coding_agent_web/internal/model"
)

const DatasetDir = "/root/coding_dataset"

func LoadKnowledgeBase() {
	files, err := os.ReadDir(DatasetDir)
	if err != nil {
		log.Println("Warning: Failed to read dataset directory:", err)
		return
	}

	count := 0
	for _, f := range files {
		if f.IsDir() || strings.HasSuffix(f.Name(), ".part") {
			continue
		}
		nameLower := strings.ToLower(f.Name())
		if strings.HasSuffix(nameLower, ".pdf") || strings.HasSuffix(nameLower, ".docx") {
			count++
		}
	}
	log.Printf("Knowledge base configured with %d dataset files from %s.", count, DatasetDir)
}

func GetDocList() []model.DocItem {
	files, err := os.ReadDir(DatasetDir)
	if err != nil {
		return nil
	}

	var docs []model.DocItem
	for _, f := range files {
		if f.IsDir() || strings.HasSuffix(f.Name(), ".part") {
			continue
		}
		name := f.Name()
		nameLower := strings.ToLower(name)
		if strings.HasSuffix(nameLower, ".pdf") || strings.HasSuffix(nameLower, ".docx") {
			info, err := f.Info()
			size := 0
			if err == nil {
				size = int(info.Size())
			}
			docs = append(docs, model.DocItem{
				Filename:  name,
				RawName:   name,
				CharCount: size,
			})
		}
	}
	return docs
}

func GetDatasetPDFListAdmin() ([]model.AdminPDFItem, error) {
	files, err := os.ReadDir(DatasetDir)
	if err != nil {
		return nil, err
	}

	var pdfs []model.AdminPDFItem
	for _, f := range files {
		if f.IsDir() || strings.HasSuffix(f.Name(), ".part") {
			continue
		}
		name := f.Name()
		nameLower := strings.ToLower(name)
		if strings.HasSuffix(nameLower, ".pdf") || strings.HasSuffix(nameLower, ".docx") {
			info, err := f.Info()
			var size int64 = 0
			if err == nil {
				size = info.Size()
			}

			pdfs = append(pdfs, model.AdminPDFItem{
				Filename:  name,
				RawName:   name,
				SizeBytes: size,
				ModTime:   info.ModTime().Format("2006-01-02 15:04"),
			})
		}
	}
	return pdfs, nil
}

func DeleteDatasetPDFAdmin(filename string) error {
	cleanName := filepath.Base(filename)
	datasetPath := filepath.Join(DatasetDir, cleanName)
	_ = os.Remove(datasetPath)
	return nil
}

func GetOriginalFilePath(rawName string) (string, string, error) {
	cleanName := filepath.Base(rawName)
	fullPath := filepath.Join(DatasetDir, cleanName)

	info, err := os.Stat(fullPath)
	if err != nil || info.IsDir() {
		return "", "", fmt.Errorf("File %s not found in dataset", cleanName)
	}

	mimeType := "application/pdf"
	if strings.HasSuffix(strings.ToLower(cleanName), ".docx") {
		mimeType = "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	}
	return fullPath, mimeType, nil
}
