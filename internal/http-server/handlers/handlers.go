package handlers

import (
	"archive/zip"
	"fmt"
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
	StartCleanup()
}

type Handler struct {
	apiEmailService ApiEmailService
	app             config.App
}

// Creating a Routine Handler
func New(config config.App) *Handler {
	return &Handler{
		app: config,
	}
}

// Generating a link to retrieve an archive by specific mail
func (h *Handler) NewUsernameEmail(w http.ResponseWriter, r *http.Request) {
	h.app.Config.Mu.Lock()
	defer h.app.Config.Mu.Unlock()
	username := wpsev.GetParam(r, "username")
	uuid := generateUUID()
	h.app.Logger.Infof("\nURL: %s \nMethod: %s \nUsername: %s \nUUID: %s", r.URL.Path, r.Method, username, uuid)
	currentTime := time.Now()
	newTime := currentTime.Add(h.app.Config.TTL * time.Second)
	h.app.LinkMap[uuid] = make(map[string]time.Time)
	h.app.LinkMap[uuid][username] = newTime
	w.Write([]byte("Уникальная ссылка: /get/" + uuid))
	//timer := time.NewTimer(h.app.Config.TTL * time.Second)
	//go func() {
	//	<-timer.C
	//	delete(h.app.LinkMap, uuid)
	//}()
}

// Getting all elements in the map
func (h *Handler) GetUsername(w http.ResponseWriter, r *http.Request) {
	h.app.Config.Mu.Lock()
	defer h.app.Config.Mu.Unlock()
	for key, value := range h.app.LinkMap {
		h.app.Logger.Infof("Key: %s Value: %s", key, value)
	}
}

// Receiving the archive by concrete mail
func (h *Handler) GetArchiveUsername(w http.ResponseWriter, r *http.Request) {
	h.app.Config.Mu.Lock()
	defer h.app.Config.Mu.Unlock()
	link := wpsev.GetParam(r, "link")
	h.app.Logger.Infof("\nURL: %s \nMethod: %s\nUUID: %s", r.URL.Path, r.Method, link)
	username, ok := h.app.LinkMap[link]

	if !ok {
		http.Error(w, "Invalid link", http.StatusBadRequest)
		return
	}

	var email string
	for keyLinkMap, valueLinkMap := range h.app.LinkMap {
		for key, _ := range valueLinkMap {
			for keyUsername := range username {
				if time.Now().After(h.app.LinkMap[keyLinkMap][keyUsername]) {
					delete(h.app.LinkMap, keyLinkMap)
					http.Error(w, "The lifetime of the link has expired ", http.StatusBadRequest)
					return
				} else if key == keyUsername {
					email = key
					break
				}
			}
		}
		if len(email) != 0 {
			break
		}
	}

	folderPath := filepath.Join(h.app.Config.StoragePath, email)
	fmt.Println(folderPath)

	zipName := email + ".zip"
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

// Creating uuid
func generateUUID() string {
	//uuid := make([]byte, 16)
	//_, _ = rand.Read(uuid)
	//return fmt.Sprintf("%x-%x-%x-%x-%x", uuid[0:4], uuid[4:6], uuid[6:8], uuid[8:10], uuid[10:])

	u := uuid.New()
	return u.String()
}

// Clearing elements in the map according to a certain period
func (h *Handler) StartCleanup(interval time.Duration) {
	go func() {
		for {
			h.app.Config.Mu.Lock()
			for keyLinkMap, valueLinkMap := range h.app.LinkMap {
				for key, _ := range valueLinkMap {
					if time.Now().After(h.app.LinkMap[keyLinkMap][key]) {
						delete(h.app.LinkMap, keyLinkMap)
					}
				}
			}
			h.app.Config.Mu.Unlock()
			time.Sleep(interval)
		}
	}()
}
