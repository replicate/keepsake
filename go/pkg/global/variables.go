package global

var Version = "development"     // set in Makefile
var Environment = "development" // set in Makefile

var ConfigFilenames []string = []string{"keepsake.yaml", "keepsake.yml", "replicate.yaml", "replicate.yml"}
var DeprecatedConfigFilenames []string = []string{"replicate.yaml", "replicate.yml"}
var Verbose = false
var WebURL = "https://keepsake.ai"
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
