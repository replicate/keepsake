package global

var Version = "development"     // set in Makefile
var Environment = "development" // set in Makefile

var ConfigFilename [2]string = [2]string{"replicate.yaml", "replicate.yml"}
var Verbose = false
var WebURL = "https://replicate.ai"
var Color = true
var ProjectDirectory = ""
var BugsEmail = "bugs@replicate.ai"
var SegmentKey = "MKaYmSZ2hW6P8OegI9g0sufjZeUh28g7"

func init() {
	if Environment == "development" {
		Version += "-dev"
	}

	if Environment == "production" {
		SegmentKey = "Fc5GClhPBLfevDXCdCJbYZuPQ1sujxEk"
	}
}
