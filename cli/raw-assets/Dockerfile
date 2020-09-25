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
{{ if .DevPythonPackage}}
COPY "{{.DevPythonPackage}}" /tmp/
RUN pip install "/tmp/{{.DevPythonPackage}}"
{{end}}
COPY . /code
WORKDIR /code
