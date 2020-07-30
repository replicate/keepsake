ARG BASE_IMAGE
FROM $BASE_IMAGE

ARG HAS_GPU
ENV HAS_GPU=$HAS_GPU

ENV DEBIAN_FRONTEND=noninteractive
{{if .InstallScript}}
COPY "{{.InstallScript}}" /tmp/install.sh
RUN chmod +x /tmp/install.sh && /tmp/install.sh
{{end}}{{range $i, $install := .Install}}
RUN {{$install}}{{end}}
{{if .HasPythonRequirements}}
COPY "{{.PythonRequirements}}" /tmp/requirements.txt
RUN pip install -r /tmp/requirements.txt
{{end}}
# FIXME: temporary, until this is on pypi or we find a better temporary spot
RUN pip install https://storage.googleapis.com/replicate-python-dev/replicate-0.0.4.tar.gz

COPY . /code
WORKDIR /code
