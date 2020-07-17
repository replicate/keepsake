package baseimages

const (
	DefaultRegistry = "us.gcr.io"
	DefaultProject  = "replicate"
	DefaultVersion  = "0.3"

	PyTorch    = "pytorch"
	Tensorflow = "tensorflow"

	Py27 = Python("2.7")
	Py35 = Python("3.5")
	Py36 = Python("3.6")
	Py37 = Python("3.7")
	Py38 = Python("3.8")

	CUDA11_0 = CUDA("11.0")
	CUDA10_2 = CUDA("10.2")
	CUDA10_1 = CUDA("10.1")
	CUDA10_0 = CUDA("10.0")
	CUDA9_2  = CUDA("9.2")
	CUDA9_1  = CUDA("9.1")
	CUDA9_0  = CUDA("9.0")
	CUDA8_0  = CUDA("8.0")

	CuDNN8 = CuDNN("8")
	CuDNN7 = CuDNN("7")
	CuDNN6 = CuDNN("6")
	CuDNN5 = CuDNN("5")

	Ubuntu18_04 = Ubuntu("18.04")
	Ubuntu16_04 = Ubuntu("16.04")
)
