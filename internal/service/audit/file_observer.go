package audit

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"

	"github.com/bigsm0uk/metrics-alert-server/internal/domain"
	"github.com/bigsm0uk/metrics-alert-server/internal/domain/interfaces"
	"go.uber.org/zap"
)

type FileObserver struct {
	path string
	log  *zap.Logger
}

var _ interfaces.AuditObserver = &FileObserver{}

func (o *FileObserver) GetID() string {
	return fmt.Sprintf("file-observer-%s", o.path)
}

func NewFileObserver(path string, log *zap.Logger) *FileObserver {
	logger := log.Named("audit-file-observer")
	return &FileObserver{path: path, log: logger}
}

func (o *FileObserver) Notify(message domain.AuditMessage) {
	if err := o.writeToFile(message); err != nil {
		o.log.Error("failed to write audit message to file", zap.Error(err))
		return
	}
	o.log.Info("audit message written to file", zap.String("path", o.path))
}

func (o *FileObserver) writeToFile(message domain.AuditMessage) error {
	file, err := os.OpenFile(o.path, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0o644)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	defer writer.Flush()

	encoder := json.NewEncoder(writer)
	if err := encoder.Encode(message); err != nil {
		return err
	}

	return nil
}
