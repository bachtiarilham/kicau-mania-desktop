package usecase

import (
	"context"
	"errors"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"kicau.desktop/cctv/internal/config"
	"kicau.desktop/cctv/internal/engine"
	"kicau.desktop/cctv/internal/player"
	"kicau.desktop/cctv/internal/recording"
)

type usecase struct {
	engine engine.LocalEngineGateway
	config config.ConfigUsecase
}

func New(engine engine.LocalEngineGateway, config config.ConfigUsecase) recording.RecordingUsecase {
	return &usecase{
		engine: engine,
		config: config,
	}
}

func sanitizeFilename(name string) string {
	r := strings.NewReplacer("/", "_", "\\", "_", ":", "_", "*", "_", "?", "_", "\"", "_", "<", "_", ">", "_", "|", "_")
	return r.Replace(strings.TrimSpace(name))
}

func (u *usecase) ExportRecording(ctx context.Context, id, suggestedFilename, destPath string) (string, error) {
	if !player.ValidID(id) {
		return "", errors.New("ID rekaman tidak valid")
	}
	res, err := u.engine.File(ctx, id)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	ext := ".mkv"
	dir := filepath.Dir(destPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}

	f, err := os.CreateTemp(dir, ".export-*"+ext)
	if err != nil {
		return "", err
	}
	defer os.Remove(f.Name())

	if _, err = io.Copy(f, res.Body); err != nil {
		f.Close()
		return "", err
	}
	_ = f.Sync()
	if err := f.Close(); err != nil {
		return "", err
	}

	if err := os.Rename(f.Name(), destPath); err != nil {
		return "", err
	}

	_ = u.config.SetRecordingsDir(dir)
	return destPath, nil
}

func (u *usecase) SaveSessionRecordings(ctx context.Context, sessionID, sessionName, targetDir, format string) ([]string, error) {
	if sessionID == "" {
		return nil, errors.New("ID sesi tidak boleh kosong")
	}
	if targetDir == "" {
		targetDir = u.config.GetRecordingsDir()
	}

	_ = u.config.SetRecordingsDir(targetDir)

	if sessionName == "" {
		sessionName = sessionID
	}
	safeSessionName := sanitizeFilename(sessionName)
	sessionFolder := filepath.Join(targetDir, safeSessionName)
	if err := os.MkdirAll(sessionFolder, 0755); err != nil {
		return nil, err
	}

	type recItem struct {
		ID        string `json:"id"`
		ChannelID string `json:"channel_id"`
		State     string `json:"state"`
	}
	var readyItems []recItem

	for attempt := 0; attempt < 15; attempt++ {
		var page struct {
			Items []recItem `json:"items"`
		}
		err := u.engine.Call(ctx, "GET", "/api/v1/recordings?limit=50&session_id="+url.QueryEscape(sessionID), nil, &page)
		if err == nil && len(page.Items) > 0 {
			allReady := true
			for _, it := range page.Items {
				if it.State != "READY" && it.State != "UPLOADED" {
					allReady = false
					break
				}
			}
			if allReady {
				readyItems = page.Items
				break
			}
			readyItems = page.Items
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(2 * time.Second):
		}
	}

	if len(readyItems) == 0 {
		return nil, errors.New("Tidak ada rekaman yang siap untuk sesi ini")
	}

	var savedFiles []string
	ext := ".mkv"
	if strings.ToLower(format) == "mp4" {
		ext = ".mp4"
	}

	for _, item := range readyItems {
		if item.ID == "" {
			continue
		}
		chName := item.ChannelID
		if chName == "" {
			chName = item.ID
		}
		destFile := filepath.Join(sessionFolder, chName+ext)

		res, err := u.engine.File(ctx, item.ID)
		if err != nil {
			continue
		}

		tmpFile := destFile + ".part"
		f, err := os.Create(tmpFile)
		if err != nil {
			res.Body.Close()
			continue
		}
		_, copyErr := io.Copy(f, res.Body)
		f.Close()
		res.Body.Close()
		if copyErr != nil {
			os.Remove(tmpFile)
			continue
		}
		_ = os.Rename(tmpFile, destFile)
		savedFiles = append(savedFiles, destFile)
	}

	return savedFiles, nil
}
