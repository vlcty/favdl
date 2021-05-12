// MIT License
//
// Copyright (c) 2021 Josef 'veloc1ty' Stautner
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/gorilla/mux"
)

const (
	FILENAME_TODOWNLOAD       string = "todownload.txt"
	FILENAME_ARCHIVE          string = "archive.txt"
	FILENAME_STORAGE          string = "storage.json"
	FILENAME_DOWNLOAD_COMMAND string = "download.sh"
)

type Batchlist struct {
	// Known Videos
	// Key: VideoID
	// Value: Downloaded
	Videos map[string]bool
}

func NewBatchlist() *Batchlist {
	return &Batchlist{Videos: make(map[string]bool)}
}

func (bl *Batchlist) CreateBatchFile() {
	file, _ := os.Create(FILENAME_TODOWNLOAD)

	for videoid, downloaded := range bl.Videos {
		if !downloaded {
			fmt.Fprintf(file, "https://www.youtube.com/watch?v=%s\n", videoid)
		}
	}

	file.Close()

	fmt.Println("Wrote batch file")
}

func (bl *Batchlist) AddVideoID(videoID string) bool {
	if _, found := bl.Videos[videoID]; !found {
		bl.Videos[videoID] = false
		return true
	}

	return false
}

func (bl *Batchlist) SaveToStorage() {
	file, fileError := os.Create(FILENAME_STORAGE)

	if fileError != nil {
		fmt.Println("Error writing storage file:", fileError.Error())
	} else {
		json.NewEncoder(file).Encode(bl.Videos)
		file.Close()

		fmt.Println("Wrote storage file")
	}
}

func (bl *Batchlist) LoadStorageFile() error {
	file, fileErr := os.Open(FILENAME_STORAGE)

	if fileErr != nil {
		return fileErr
	}

	defer file.Close()

	if decodeErr := json.NewDecoder(file).Decode(&bl.Videos); decodeErr != nil {
		return decodeErr
	}

	fmt.Println("Loaded storage file")

	return nil
}

func (bl *Batchlist) LoadYoutubeDLArchive() error {
	file, fileErr := os.Open(FILENAME_ARCHIVE)

	if fileErr != nil {
		return fileErr
	}

	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()

		if strings.Contains(line, "youtube ") {
			videoid := strings.ReplaceAll(line, "youtube ", "")

			if len(videoid) != 11 {
				fmt.Println("No valid videoid found")
			} else {
				bl.Videos[videoid] = true
			}
		} else {
			fmt.Println("Skipping invalid line:", line)
		}
	}

	fmt.Println("Loaded archive")

	return nil
}

var batchlist *Batchlist

func main() {
	batchlist = NewBatchlist()
	if e := batchlist.LoadStorageFile(); e != nil {
		fmt.Println(e)
		os.Exit(1)
	}

	if e := batchlist.LoadYoutubeDLArchive(); e != nil {
		fmt.Println(e)
		os.Exit(1)
	}

	r := mux.NewRouter()
	r.HandleFunc("/videos/add", VideoAddHandler).Methods(http.MethodPost)
	r.HandleFunc("/videos/add", VideoAddHandlerCORS).Methods(http.MethodOptions)
	r.HandleFunc("/createbatch", CreateBatchFile).Methods(http.MethodPost)
	r.HandleFunc("/reloadarchive", ReloadArchive).Methods(http.MethodPost)

	http.Handle("/", r)

	fmt.Println("Waiting for requests")

	if err := http.ListenAndServe("[::]:4242", r); err != nil {
		fmt.Println(err)
	}
}

func CreateBatchFile(response http.ResponseWriter, request *http.Request) {
	request.Body.Close()

	batchlist.CreateBatchFile()

	response.WriteHeader(http.StatusOK)
	fmt.Fprintln(response, "Created batch file", FILENAME_TODOWNLOAD)
}

func ReloadArchive(response http.ResponseWriter, request *http.Request) {
	request.Body.Close()

	if e := batchlist.LoadYoutubeDLArchive(); e != nil {
		fmt.Println(e)
	} else {
		fmt.Println("Reloaded archive")
	}

	batchlist.SaveToStorage()

	response.WriteHeader(http.StatusOK)
}

func VideoAddHandlerCORS(response http.ResponseWriter, request *http.Request) {
	request.Body.Close()

	response.Header().Set("Access-Control-Allow-Origin", "*")
	response.Header().Set("Access-Control-Allow-Methods", "POST")
	response.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	response.WriteHeader(http.StatusOK)
}

func VideoAddHandler(response http.ResponseWriter, request *http.Request) {
	content := make([]string, 0)
	defer request.Body.Close()

	if decodeErr := json.NewDecoder(request.Body).Decode(&content); decodeErr != nil {
		fmt.Println("Decode err:", decodeErr.Error())
		http.Error(response, decodeErr.Error(), http.StatusBadRequest)
		return
	}

	addedVideoIDs := 0

	for _, videoID := range content {
		if batchlist.AddVideoID(videoID) {
			addedVideoIDs++
		}
	}

	fmt.Println("Added", addedVideoIDs, "new video IDs")

	batchlist.SaveToStorage()
	batchlist.CreateBatchFile()

	response.WriteHeader(http.StatusOK)
}
