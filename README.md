# favdl

Download your liked videos from YouTube in the most complicated way possible!

Requirements:

- Working youtube-dl installation
- Linux
- Must know how hot open the developer tools in your browser (tested only with Firefox)

## How it works

The application:
- Spawns a webserver and waits for REST requests
- Loads all previously known video IDs the file storage.json
- Reads the youtube-dl archive with already downloaded videos (to prevent unnecessary calls to YouTube) and processes it
- Creates a batch file for youtube-dl with links to missing videos on request

The browser / The JavaScript:
- Navigate to your liked videos. The URL is: https://www.youtube.com/playlist?list=LL
- The java script code you need to paste into the console works on the DOM and extracts all videolinks
- All extraced video links are sent to the application via a REST call
- Pro tip: YouTube only loads 100 Videos. So if you do this initially scroll to the bottom and THEN execute the js code. If you only want to "sync" the new ones 100 should be enough.

youtube-dl:
- Downloads all videos from the batchfile
- Stores already downloaded videos into the archive.txt file

## Usage

1) Close this repo
2) Create empty files: `touch archive.txt todownload.txt storage.json`
3) Start the application: `go run favdl.go`
4) Copy the content from browsercode.js into the java script console of your browser
5) A log message shoud appear. Something like "Added X new video ids".
6) Issue a POST request to the application: `curl -X POST localhost:4242/createbatch`
7) todownload.txt should now contain all missing videos
8) Issue the youtube-dl command from below
9) When the download is finished reload the archive: `curl -X POST localhost:4242/reloadarchive`

If you want to download new videos start from step 3. Easy, isn't it?

## youtube-dl command

The basic command you need to use:

> youtube-dl --batch-file todownload.txt --download-archive archive.txt -o "downloads/%(title)s.%(ext)s"

Add more parameters as required:

- "--merge-output-format mkv" -> Convert all videos to mkv
- "--limit-rate 5M" -> Limit bandwidth to 5 MByte/s

## Known limitations

- I've only tested it in Firefox and on Linux/Mac. If you have Windows GTFO.
- If you get a "bind error" on application startup you most likely have disabled IPv6 and you should be ashamed of yourself
- This works until browser vendors are blocking calls to unsecured endpoints from secured sites (HTTPS -> HTTP). Currently the application does not provide a HTTPS secured endpoint.

## Automation

You can automate it:

- Compile the go code to a static binay and upload it somewhere
- Write a systemd service for it. Don't forget to set "WorkingDirectory"!
- Touch the needed files as described above
- Download https://addons.mozilla.org/de/firefox/addon/codeinjector/ and add a rule to inject the content from "browsercode.js". Don't forget to adapt the BACKEND_URL!

## Last words

This project was coded on a single evening. It's officially in the "works for me" and "works on my machine" category. It lacks basic error and safety checks. You should never rely onto it.
