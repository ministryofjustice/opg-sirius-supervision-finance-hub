# Use an official Go image
FROM golang:1.26@sha256:f96cc555eb8db430159a3aa6797cd5bae561945b7b0fe7d0e284c63a3b291609

# Install dependencies
RUN apt-get update && apt-get install -y \
    curl \
    unzip \
    libssl-dev \
    && rm -rf /var/lib/apt/lists/*

# Install pact-ffi
RUN curl -LO https://github.com/pact-foundation/pact-ffi/releases/latest/download/libpact_ffi-linux-x86_64.tar.gz \
    && tar -xzf libpact_ffi-linux-x86_64.tar.gz \
    && mv libpact_ffi.so /usr/local/lib/ \
    && ldconfig

# Set up working directory
WORKDIR /app

# Run tests
CMD ["go", "test", "-v", "./..."]
