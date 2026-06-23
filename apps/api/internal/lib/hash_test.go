package lib

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHashBytes_Success(t *testing.T) {
	data := []byte("test data")

	hash := HashBytes(data)

	assert.NotEmpty(t, hash)
	assert.Len(t, hash, 64) // SHA-256 produces 64 hex characters
}

func TestHashBytes_Consistency(t *testing.T) {
	data := []byte("test data")

	hash1 := HashBytes(data)
	hash2 := HashBytes(data)

	assert.Equal(t, hash1, hash2)
}

func TestHashBytes_Different(t *testing.T) {
	data1 := []byte("data 1")
	data2 := []byte("data 2")

	hash1 := HashBytes(data1)
	hash2 := HashBytes(data2)

	assert.NotEqual(t, hash1, hash2)
}

func TestHashFile_Success(t *testing.T) {
	data := []byte("test file content")
	reader := bytes.NewReader(data)

	hash, err := HashFile(reader)

	assert.NoError(t, err)
	assert.NotEmpty(t, hash)
	assert.Len(t, hash, 64)
}

func TestHashFile_Consistency(t *testing.T) {
	data := []byte("test file content")

	reader1 := bytes.NewReader(data)
	hash1, _ := HashFile(reader1)

	reader2 := bytes.NewReader(data)
	hash2, _ := HashFile(reader2)

	assert.Equal(t, hash1, hash2)
}

func TestHashFile_Different(t *testing.T) {
	data1 := []byte("file 1")
	data2 := []byte("file 2")

	reader1 := bytes.NewReader(data1)
	hash1, _ := HashFile(reader1)

	reader2 := bytes.NewReader(data2)
	hash2, _ := HashFile(reader2)

	assert.NotEqual(t, hash1, hash2)
}
