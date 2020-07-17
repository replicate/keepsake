FROM {{.BaseImage}}

ENV DEBIAN_FRONTEND=noninteractive

# If you add something to this list, add a comment to explain what it's for.
# It's often not clear what long lists of  apt packages in Dockerfiles are used for,
# and this will help us keep it tidy.
RUN apt-get update -q && apt-get install -qy --no-install-recommends \
        make \
        build-essential \
        libssl-dev \
        zlib1g-dev \
        libbz2-dev \
        libreadline-dev \
        libsqlite3-dev \
        wget \
        curl \
        llvm \
        libncurses5-dev \
        libncursesw5-dev \
        xz-utils \
        tk-dev \
        libffi-dev \
        liblzma-dev \
        python-openssl \
        git \
        ca-certificates \
        dtrx \
        && rm -rf /var/lib/apt/lists/*

RUN curl https://pyenv.run | bash
ENV PATH="/root/.pyenv/shims:/root/.pyenv/bin:$PATH"
RUN git clone https://github.com/momo-lab/pyenv-install-latest.git "$(pyenv root)"/plugins/pyenv-install-latest && \
        pyenv install-latest {{.PythonVersion}}
RUN pyenv global $(pyenv install-latest --print {{.PythonVersion}})

# TODO: pin version (jupyter is a weird meta-package which can't be pinned)
RUN pip install jupyter
