const BACKEND_URL = "http://[::1]:4242/videos/add"

// Only run when we lookt at our liked playlist
if ( document.location == "https://www.youtube.com/playlist?list=LL" ) {
    list = document.getElementsByClassName("yt-simple-endpoint style-scope ytd-playlist-video-renderer")

    ytwatchurl = "https://www.youtube.com/watch"

    foundVideoIDs = []

    Object.keys(list).forEach(key => {
        // Check if the element is from the expected type
        if ( list[key].nodeName == "A" && list[key].id == "video-title" && list[key].href.includes(ytwatchurl) ) {
            urlParams = new URLSearchParams(list[key].href.replace(ytwatchurl, ""))

            // Does the URL contain a video ID?
            if (urlParams.has("v")) {
                foundVideoIDs.push(urlParams.get("v"))
            }
        }
    })

    if ( foundVideoIDs.length > 0 ) {
        console.log("Found", foundVideoIDs.length, "video IDs")
        console.log(foundVideoIDs)

        var xhr = new XMLHttpRequest()
        xhr.open("POST", BACKEND_URL, true)
        xhr.setRequestHeader('Content-Type', 'application/json')
        xhr.send(JSON.stringify(foundVideoIDs))
    }
}
