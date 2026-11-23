package domain

import (
	"context"
)

// RecordingUsecase handles export operations and session auto-saving with human-readable file naming.
type RecordingUsecase interface {
	ExportRecording(ctx context.Context, id, suggestedFilename, destPath string) (string, error)
	SaveSessionRecordings(ctx context.Context, sessionID, sessionName, targetDir, format string) ([]string, error)
}
