package util

import (
	"bytes"
	"compress/gzip"
	"io"

	"github.com/goccy/go-json"
)

type Number interface {
	int64 |
		float64 | int | float32
}

func GetDefault[T Number](value *T) T {
	if value == nil {
		return 0
	}
	return *value
}

// CompressJSON сжимает данные в JSON формате с использованием gzip
func CompressJSON(data []byte) ([]byte, error) {

	var buf bytes.Buffer

	writer := gzip.NewWriter(&buf)

	_, err := writer.Write(data)
	if err != nil {
		writer.Close()
		return nil, err
	}

	err = writer.Close()
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// DecompressJSON декомпрессирует gzip данные и преобразует в JSON
func DecompressJSON(compressedData []byte, target any) error {
	reader := bytes.NewReader(compressedData)

	gzipReader, err := gzip.NewReader(reader)
	if err != nil {
		return err
	}
	defer gzipReader.Close()

	decompressedData, err := io.ReadAll(gzipReader)
	if err != nil {
		return err
	}

	return json.Unmarshal(decompressedData, target)
}
