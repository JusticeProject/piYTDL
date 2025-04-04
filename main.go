package main

import (
	"fmt"
	"html/template"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

///////////////////////////////////////////////////////////////////////////////////////////////////

type templateInfo struct {
	MSG      string
	ID       string
	FILENAME string
}

var templates = template.Must(template.ParseFiles("templates/downloader.html",
	"templates/inprogress.html", "templates/finished.html", "templates/failed.html"))

type statusInfo struct {
	finished	bool
	failed		bool
	converting	bool
	filename	string
	errMsg		string
}

// the key for the map is the id for the client
var statusMap = make(map[string]statusInfo)
var mutex sync.Mutex

///////////////////////////////////////////////////////////////////////////////////////////////////

func getDownloadFolder(id string) string {
	if len(id) > 0 {
		return filepath.Join(".", "downloads", id)
	} else {
		return filepath.Join(".", "downloads")
	}
}

///////////////////////////////////////////////////////////////////////////////////////////////////

func addToStatusMap(id string) {
	mutex.Lock()
	statusMap[id] = statusInfo{finished: false, failed: false, converting: false}
	mutex.Unlock()
}

///////////////////////////////////////////////////////////////////////////////////////////////////

func addFilename(id string, filename string) {
	mutex.Lock()
	status := statusMap[id]
	status.filename = filename
	statusMap[id] = status
	mutex.Unlock()
}

///////////////////////////////////////////////////////////////////////////////////////////////////

func removeFromStatusMap(id string) {
	mutex.Lock()
	delete(statusMap, id)
	mutex.Unlock()
}

///////////////////////////////////////////////////////////////////////////////////////////////////

func markIdConverting(id string) {
	mutex.Lock()
	status, exists := statusMap[id]
	if exists {
		status.converting = true
		statusMap[id] = status
	} else {
		statusMap[id] = statusInfo{finished: false, failed: true, converting: false}
	}
	mutex.Unlock()
}

///////////////////////////////////////////////////////////////////////////////////////////////////

func markIdFinished(id string) {
	mutex.Lock()
	status, exists := statusMap[id]
	if exists {
		status.finished = true
		statusMap[id] = status
	} else {
		statusMap[id] = statusInfo{finished: false, failed: true, converting: false}
	}
	mutex.Unlock()
}

///////////////////////////////////////////////////////////////////////////////////////////////////

func markIdFailed(id string, errMsg string) {
	mutex.Lock()
	status, exists := statusMap[id]
	if exists {
		status.failed = true
		status.errMsg = errMsg
		statusMap[id] = status
	} else {
		statusMap[id] = statusInfo{finished: false, failed: true, converting: false, errMsg: errMsg}
	}
	mutex.Unlock()
}

///////////////////////////////////////////////////////////////////////////////////////////////////

func getIdStatus(id string) statusInfo {
	mutex.Lock()
	status, exists := statusMap[id]
	mutex.Unlock()

	if exists {
		return status
	} else {
		return statusInfo{finished: false, failed: true, converting: false}
	}
}

///////////////////////////////////////////////////////////////////////////////////////////////////

func cleanOldFiles(forceDelete bool) {
	downloadFolder := getDownloadFolder("")

	entries, _ := os.ReadDir(downloadFolder)
	for _, entry := range entries {
		if entry.IsDir() {
			folderInfo, _ := entry.Info()

			modifiedTime := folderInfo.ModTime()
			currentTime := time.Now()
			minutesAgo := currentTime.Sub(modifiedTime).Minutes()

			// remove old files older than 3 hours
			if forceDelete || minutesAgo > 180 {
				relativePath := filepath.Join(downloadFolder, entry.Name())
				fmt.Println("deleting old folder", relativePath)
				os.RemoveAll(relativePath)

				// also remove old entries in the status map that are tied to the folder name/id
				removeFromStatusMap(entry.Name())
			}
		}
	}
}

///////////////////////////////////////////////////////////////////////////////////////////////////

func cleanupThread() {
	firstTime := true

	for {
		cleanOldFiles(firstTime)
		firstTime = false

		time.Sleep(20 * time.Minute)
	}
}

///////////////////////////////////////////////////////////////////////////////////////////////////

func downloaderThread(id string, url string, format string) {
	fmt.Println("downloading", url, format)

	// create the folder where the client will grab it
	folderPath := getDownloadFolder(id)
	err := os.MkdirAll(folderPath, 0777)
	if err != nil {
		markIdFailed(id, err.Error())
		fmt.Println(err)
		return
	}

	fmt.Println("created folder", folderPath)

	// download video to folderPath from above
	outBytes, err := exec.Command("bash", "helper.sh", url, folderPath).CombinedOutput()
	if err != nil {
		fmt.Println("helper.sh returned an error:", err)
		fmt.Printf("additional info: %s\n", outBytes)
		markIdFailed(id, string(outBytes))
		return
	}

	fmt.Println("download complete")

	videoFilename := ""
	outString := string(outBytes[:])
	lines := strings.Split(outString, "\n")
	for _, line := range lines {
		if strings.Contains(line, "Merging formats") && strings.Contains(line, "\"") {
			videoFilename = strings.Split(line, "\"")[1]
			// could be .mp4, .webm, .mkv, or other
		}
	}

	// if we didn't find the video filename it may not have downloaded, don't continue
	if videoFilename == "" {
		fmt.Println("videoFilename not found")
		markIdFailed(id, "videoFilename not found")
		return
	}

	fmt.Println("found videoFilename", videoFilename)

	// remove illegal chars from filename first
	newVideoFilename := removeIllegalChars(videoFilename)
	if newVideoFilename != videoFilename {
		err = os.Rename(videoFilename, newVideoFilename)
		if err != nil {
			fmt.Println("Could not rename video file:", err)
			markIdFailed(id, err.Error())
			return
		}
		videoFilename = newVideoFilename
	}

	fmt.Println("cleaned up videoFilename", videoFilename)

	if format == "audio" {
		markIdConverting(id)
		
		extension := path.Ext(videoFilename)
		audioFilename := strings.TrimSuffix(videoFilename, extension) + ".mp3"
		fmt.Println("using audioFilename", audioFilename)

		// ffmpeg -i "downloads/id/blah.webm" -b:a 128k "downloads/id/blah.mp3"
		outBytes, err = exec.Command("ffmpeg", "-i", videoFilename, "-b:a", "128k", audioFilename).CombinedOutput()
		if err != nil {
			fmt.Println("ffmpeg returned error:", err)
			fmt.Printf("additional info: %s\n", outBytes)
			markIdFailed(id, string(outBytes))
			return
		}

		fmt.Println("ffmpeg complete, removing video")

		err = os.Remove(videoFilename)
		if err != nil {
			fmt.Println("os.Remove returned", err)
		}
		// will be split into a max of 3 substrings
		nameSplit := strings.SplitN(audioFilename, "/", 3)
		audioFilename = nameSplit[len(nameSplit)-1]
		fmt.Println("new audioFilename is", audioFilename)
		addFilename(id, audioFilename)
	} else {
		// will be split into a max of 3 substrings
		nameSplit := strings.SplitN(videoFilename, "/", 3)
		videoFilename = nameSplit[len(nameSplit)-1]
		addFilename(id, videoFilename)
	}

	markIdFinished(id)
}

///////////////////////////////////////////////////////////////////////////////////////////////////

func onHomePage(w http.ResponseWriter, req *http.Request) {
	err := templates.ExecuteTemplate(w, "downloader.html", nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

///////////////////////////////////////////////////////////////////////////////////////////////////

func onDownloadRequest(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Redirect(w, req, "/failed.html", http.StatusFound)
		return
	}

	url := req.FormValue("youtubeurl")
	format := req.FormValue("format")
	id := randomString()

	// we won't check for valid url here, that will be done in downloaderThread

	addToStatusMap(id)
	go downloaderThread(id, url, format)

	http.Redirect(w, req, "/inprogress.html?id="+id, http.StatusFound)
}

///////////////////////////////////////////////////////////////////////////////////////////////////

func onInProgress(w http.ResponseWriter, req *http.Request) {
	params := req.URL.Query()
	id := params.Get("id")

	status := getIdStatus(id)

	if status.failed {
		http.Redirect(w, req, "/failed.html?id="+id, http.StatusFound)
	} else if status.finished {
		http.Redirect(w, req, "/finished.html?id="+id, http.StatusFound)
	} else {
		// send them to the same page, they need to keep waiting
		msg := "Downloading..."

		if status.converting {
			msg += "done. Converting..."
		}
		
		err := templates.ExecuteTemplate(w, "inprogress.html", templateInfo{MSG: msg})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

///////////////////////////////////////////////////////////////////////////////////////////////////

func onFinished(w http.ResponseWriter, req *http.Request) {
	params := req.URL.Query()
	id := params.Get("id")
	status := getIdStatus(id)

	err := templates.ExecuteTemplate(w, "finished.html", templateInfo{ID: id, FILENAME: status.filename})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

///////////////////////////////////////////////////////////////////////////////////////////////////

func onGetFile(w http.ResponseWriter, req *http.Request) {
	// TODO: check for ".."
	// TODO: check for 3 '/' chars
	// url will be in the format /getfile/id/filename
	urlSplit := strings.Split(req.URL.Path, "/")
	id := urlSplit[2]
	fmt.Println("file requested for id", id)

	folder := getDownloadFolder(id)
	entries, err := os.ReadDir(folder)

	if err != nil {
		err := templates.ExecuteTemplate(w, "failed.html?id="+id, templateInfo{MSG: err.Error()})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			// there should only be one file, so if this is a file then send to client
			relativePath := filepath.Join(folder, entry.Name())
			fmt.Println("serving", relativePath)
			http.ServeFile(w, req, relativePath)
			// remove the file and folder now that it has been sent to client
			os.RemoveAll(folder)
			removeFromStatusMap(id)
			// in case there is more than one file, return right away
			return
		}
	}
}

///////////////////////////////////////////////////////////////////////////////////////////////////

func onFailed(w http.ResponseWriter, req *http.Request) {
	params := req.URL.Query()
	id := params.Get("id")
	errMsg := getIdStatus(id).errMsg
	err := templates.ExecuteTemplate(w, "failed.html", templateInfo{MSG: errMsg})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

///////////////////////////////////////////////////////////////////////////////////////////////////

func main() {
	// create downloads folder if it doesn't already exist
	err := os.MkdirAll("downloads", 0777)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// removes old files and old entries in the statusMap
	go cleanupThread()

	// determine our local IP address and port number, keep trying if failed
	localIP := ""
	for localIP == "" {
		localIP, err = getLocalIP()
		if err != nil {
			fmt.Println(err)
			time.Sleep(time.Second)
		}
		
	}
	localIP += ":6514" // add the port number
	fmt.Println("web server will run on", localIP)

	// set up all of the URL endpoints
	http.HandleFunc("/downloader.html", onHomePage)
	http.HandleFunc("/downloadrequest", onDownloadRequest)
	http.HandleFunc("/inprogress.html", onInProgress)
	http.HandleFunc("/finished.html", onFinished)
	http.HandleFunc("/getfile/", onGetFile)
	http.HandleFunc("/failed.html", onFailed)

	// start serving
	err = http.ListenAndServe(localIP, nil)
	if err != nil {
		fmt.Println("Unable to serve urls on address", localIP)
		os.Exit(1)
	}
}
