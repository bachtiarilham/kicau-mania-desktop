package recording

import "context"

type RecordingUsecase interface {
	ExportRecording(ctx context.Context, id, suggestedFilename, destPath string) (string, error)
	SaveSessionRecordings(ctx context.Context, sessionID, sessionName, targetDir, format string) ([]string, error)
}
