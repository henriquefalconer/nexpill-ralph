FROM ubuntu:24.04

RUN apt-get update && apt-get install -y --no-install-recommends \
        bash ca-certificates curl git openssh-client sudo \
    && rm -rf /var/lib/apt/lists/*

RUN git config --global --add safe.directory '*'

RUN curl -fsSL https://claude.ai/install.sh | bash
ENV PATH="/root/.local/bin:${PATH}"

RUN curl -fsSL https://go.dev/dl/go1.22.3.linux-amd64.tar.gz \
      | tar -C /usr/local -xz
ENV PATH="/usr/local/go/bin:${PATH}"

WORKDIR /workspace
