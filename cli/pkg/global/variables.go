package global

var Version = "development" // set in Makefile
var Verbose = false
var WebURL = "https://beta2.replicate.ai"
var Color = true
var SourceDirectory = ""
var BugsEmail = "bugs@replicate.ai"
var ReplicateDownloadURLs = map[string]string{
	"linux":   "https://storage.googleapis.com/replicate-public/cli/latest/linux/amd64/replicate",
	"windows": "https://storage.googleapis.com/replicate-public/cli/latest/windows/amd64/replicate",
	"darwin":  "https://storage.googleapis.com/replicate-public/cli/latest/darwin/amd64/replicate",
}
