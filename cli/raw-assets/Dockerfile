ARG BASE_IMAGE
FROM $BASE_IMAGE

ARG HAS_GPU
ENV HAS_GPU=$HAS_GPU

ENV DEBIAN_FRONTEND=noninteractive
{{range $i, $install := .Install}}
RUN {{$install}}{{end}}
{{if .HasPythonRequirements}}
COPY {{.PythonRequirements}} /tmp/requirements.txt
RUN pip install -r /tmp/requirements.txt
{{end}}
# FIXME: temporary, until this is on pypi or we find a better temporary spot
RUN pip install https://storage.googleapis.com/bfirsh-dev/replicate-python/replicate-0.0.1.tar.gz

COPY . /code
WORKDIR /code
