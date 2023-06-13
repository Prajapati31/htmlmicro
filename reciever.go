package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
)

const (
	chunkSize = 1024 * 1024 // 1 MB
	tempDir   = "temp"
)

func main() {
	http.HandleFunc("/receive", receiveHandler)
	log.Fatal(http.ListenAndServe(":8081", nil))
}

func receiveHandler(w http.ResponseWriter, r *http.Request) {
	chunk, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fileName := r.Header.Get("FileName")
	err = saveChunk(fileName, chunk)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprint(w, "Chunk received successfully")
}

func saveChunk(fileName string, chunk []byte) error {
	err := os.MkdirAll(tempDir, os.ModePerm)
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %v", err)
	}

	chunkPath := filepath.Join(tempDir, fileName)
	f, err := os.OpenFile(chunkPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to save chunk: %v", err)
	}
	defer f.Close()

	_, err = f.Write(chunk)
	if err != nil {
		return fmt.Errorf("failed to write chunk: %v", err)
	}

	return nil
}

func assembleChunks(tempDir string, fileName string) error {
	chunkFiles, err := ioutil.ReadDir(tempDir)
	if err != nil {
		return fmt.Errorf("failed to read chunk directory: %v", err)
	}

	sortChunkFiles(chunkFiles)

	finalPath := filepath.Join(".", fileName)
	finalFile, err := os.Create(finalPath)
	if err != nil {
		return fmt.Errorf("failed to create final file: %v", err)
	}
	defer finalFile.Close()

	for _, chunkFile := range chunkFiles {
		chunkPath := filepath.Join(tempDir, chunkFile.Name())
		chunkData, err := ioutil.ReadFile(chunkPath)
		if err != nil {
			return fmt.Errorf("failed to read chunk file: %v", err)
		}

		_, err = finalFile.Write(chunkData)
		if err != nil {
			return fmt.Errorf("failed to write chunk to final file: %v", err)
		}
	}

	err = os.RemoveAll(tempDir)
	if err != nil {
		return fmt.Errorf("failed to remove temp directory: %v", err)
	}

	return nil
}

func sortChunkFiles(chunkFiles []os.FileInfo) {
	sort.Slice(chunkFiles, func(i, j int) bool {
		return chunkFiles[i].Name() < chunkFiles[j].Name()
	})
}
