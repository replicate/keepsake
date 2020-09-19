FROM golang:1.14-buster
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
RUN pyenv install 3.8.5
RUN pyenv global 3.8.5
