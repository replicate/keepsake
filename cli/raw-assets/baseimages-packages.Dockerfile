FROM {{.BaseImage}}

RUN pip install {{.PythonPackages}}
