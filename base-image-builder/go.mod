module replicate.ai/image-builder

go 1.13

require (
	cloud.google.com/go/storage v1.6.0
	github.com/apex/log v1.1.2
	github.com/dchest/uniuri v0.0.0-20200228104902-7aecb25e1fe5
	github.com/dsnet/compress v0.0.1 // indirect
	github.com/frankban/quicktest v1.10.0 // indirect
	github.com/golang/snappy v0.0.1 // indirect
	github.com/mholt/archiver v3.1.1+incompatible
	github.com/nwaples/rardecode v1.1.0 // indirect
	github.com/pierrec/lz4 v2.5.2+incompatible // indirect
	github.com/spf13/cobra v1.0.0
	github.com/ulikunitz/xz v0.5.7 // indirect
	github.com/xi2/xz v0.0.0-20171230120015-48954b6210f8 // indirect
	google.golang.org/api v0.29.0
	replicate.ai/cli v0.0.0
)

replace replicate.ai/cli v0.0.0 => ../cli
