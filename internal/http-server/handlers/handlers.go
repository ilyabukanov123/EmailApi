package handlers

import (
	"archive/zip"
	"github.com/google/uuid"
	"github.com/ilyabukanov123/api-mail/internal/config"
	"github.com/ilyabukanov123/api-mail/internal/lib/wpsev"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

type ApiEmailService interface {
	NewUsernameEmail(w http.ResponseWriter, r *http.Request) (string, error)
	GetUsername()
	GetArchiveUsername()
}

type Handler struct {
	apiEmailService ApiEmailService
	app             config.App
}

func New(config config.App) *Handler {
	return &Handler{
		app: config,
	}
}

func (h *Handler) NewUsernameEmail(w http.ResponseWriter, r *http.Request) {
	username := wpsev.GetParam(r, "username")
	uuid := generateUUID()
	h.app.Logger.Infof("\nURL: %s \nMethod: %s \nUsername: %s \nUUID: %s", r.URL.Path, r.Method, username, uuid)
	h.app.LinkMap[uuid] = username
	w.Write([]byte("Уникальная ссылка: /get/" + uuid))
	timer := time.NewTimer(h.app.Config.TTL * time.Second)
	go func() {
		<-timer.C
		delete(h.app.LinkMap, uuid)
	}()
}

func (h *Handler) GetUsername(w http.ResponseWriter, r *http.Request) {
	for key, value := range h.app.LinkMap {
		h.app.Logger.Infof("Key: %s Value: %s", key, value)
	}
}

func (h *Handler) GetArchiveUsername(w http.ResponseWriter, r *http.Request) {
	link := wpsev.GetParam(r, "link")
	h.app.Logger.Infof("\nURL: %s \nMethod: %s\nUUID: %s", r.URL.Path, r.Method, link)
	username, ok := h.app.LinkMap[link]
	if !ok {
		http.Error(w, "Invalid link", http.StatusBadRequest)
		return
	}

	folderPath := filepath.Join(h.app.Config.StoragePath, username)

	zipName := username + ".zip"
	zipPath := filepath.Join(h.app.Config.StoragePath, zipName)
	zipFile, err := os.Create(zipPath)
	if err != nil {
		h.app.Logger.Errorf("Failed to create archive %s", err)
		return
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	err = filepath.Walk(folderPath, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			h.app.Logger.Errorf("Failed to access the file or folder %s", err)
		}

		relativePath, err := filepath.Rel(folderPath, filePath)
		if err != nil {
			h.app.Logger.Errorf("Failed to calculate the relative path of a file or folder to create the appropriate folder structure within the archive %s", err)
		}

		zipFile, err := zipWriter.Create(relativePath)
		if err != nil {
			h.app.Logger.Errorf("Failed to create a new file inside the archive %s", err)
			return err
		}

		if !info.IsDir() {
			file, err := os.Open(filePath)
			if err != nil {
				h.app.Logger.Errorf("Unable to open file %s", err)
				return err
			}
			defer file.Close()

			_, err = io.Copy(zipFile, file)
			if err != nil {
				h.app.Logger.Errorf("Failed to copy the contents of the file in the folder to the created file in the archive  %s", err)
				return err
			}
		}

		return nil
	})

	pdfFile, err := os.Open(h.app.Config.StoragePath + "/readme.pdf")
	if err != nil {
		h.app.Logger.Errorf("Failed to open PDF file: %s", err)
	}
	defer pdfFile.Close()

	pdfWriter, err := zipWriter.Create("readme.pdf")
	if err != nil {
		h.app.Logger.Errorf("Failed to create PDF file inside the archive: %s", err)
	}

	_, err = io.Copy(pdfWriter, pdfFile)
	if err != nil {
		h.app.Logger.Errorf("Failed to copy PDF file contents to the archive: %s", err)
	}

	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", "attachment; filename="+zipName)
	http.ServeFile(w, r, zipPath)

	err = os.Remove(zipPath)
	if err != nil {
		h.app.Logger.Errorf("Failed to remove archive: %s", err)
	}
}

func generateUUID() string {
	//uuid := make([]byte, 16)
	//_, _ = rand.Read(uuid)
	//return fmt.Sprintf("%x-%x-%x-%x-%x", uuid[0:4], uuid[4:6], uuid[6:8], uuid[8:10], uuid[10:])

	u := uuid.New()
	return u.String()
}
