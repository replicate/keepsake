package global

var Version = "development"     // set in Makefile
var Environment = "development" // set in Makefile

var ConfigFilename = "replicate.yaml"
var Verbose = false
var WebURL = "https://replicate.ai"
var Color = true
var ProjectDirectory = ""
var BugsEmail = "bugs@replicate.ai"
var ReplicateDownloadURLs = map[string]string{
	"linux":   "https://storage.googleapis.com/replicate-public/cli/latest/linux/amd64/replicate",
	"windows": "https://storage.googleapis.com/replicate-public/cli/latest/windows/amd64/replicate",
	"darwin":  "https://storage.googleapis.com/replicate-public/cli/latest/darwin/amd64/replicate",
}
var SegmentKey = "MKaYmSZ2hW6P8OegI9g0sufjZeUh28g7"

func init() {
	if Environment == "development" {
		Version += "-dev"
	}

	if Environment == "production" {
		SegmentKey = "Fc5GClhPBLfevDXCdCJbYZuPQ1sujxEk"
	}
}
