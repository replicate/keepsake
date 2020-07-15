module replicate.ai/image-builder

go 1.13

require (
	cloud.google.com/go v0.38.0
	github.com/apex/log v1.1.2
	github.com/dchest/uniuri v0.0.0-20200228104902-7aecb25e1fe5
	github.com/dsnet/compress v0.0.1 // indirect
	github.com/frankban/quicktest v1.10.0 // indirect
	github.com/golang/snappy v0.0.1 // indirect
	github.com/googleapis/gax-go v2.0.2+incompatible // indirect
	github.com/mholt/archiver v3.1.1+incompatible
	github.com/nwaples/rardecode v1.1.0 // indirect
	github.com/pierrec/lz4 v2.5.2+incompatible // indirect
	github.com/spf13/cobra v1.0.0
	github.com/ulikunitz/xz v0.5.7 // indirect
	github.com/xi2/xz v0.0.0-20171230120015-48954b6210f8 // indirect
	go.opencensus.io v0.22.3 // indirect
	golang.org/x/oauth2 v0.0.0-20190604053449-0f29369cfe45
	golang.org/x/sync v0.0.0-20190911185100-cd5d95a43a6e
	google.golang.org/api v0.22.0
	replicate.ai/cli v0.0.0
)

replace replicate.ai/cli v0.0.0 => ../cli
