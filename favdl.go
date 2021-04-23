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
}

func (bl *Batchlist) AddVideoID(videoID string) bool {
	if _, found := bl.Videos[videoID]; !found {
		bl.Videos[videoID] = false
		return true
	}

	return false
}

func (bl *Batchlist) SaveToStorage() {
	file, _ := os.Create(FILENAME_STORAGE)
	json.NewEncoder(file).Encode(bl.Videos)
	file.Close()
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

			if _, found := bl.Videos[videoid]; found {
				// Mark as downloaded
				bl.Videos[videoid] = true
			}
		}
	}

	return nil
}

var batchlist *Batchlist

func main() {
	batchlist = NewBatchlist()
	if e := batchlist.LoadStorageFile(); e != nil {
		fmt.Println(e)
	}

	if e := batchlist.LoadYoutubeDLArchive(); e != nil {
		fmt.Println(e)
	}

	r := mux.NewRouter()
	r.HandleFunc("/videos/add", VideoAddHandler).Methods(http.MethodPost)
	r.HandleFunc("/videos/add", VideoAddHandlerCORS).Methods(http.MethodOptions)
	r.HandleFunc("/createbatch", CreateBatchFile).Methods(http.MethodPost)
	r.HandleFunc("/reloadarchive", ReloadArchive).Methods(http.MethodPost)

	http.Handle("/", r)
	r.Use(mux.CORSMethodMiddleware(r))

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

	batchlist.SaveToStorage()
	fmt.Println("Added", addedVideoIDs, "new video IDs")

	response.WriteHeader(http.StatusOK)
}