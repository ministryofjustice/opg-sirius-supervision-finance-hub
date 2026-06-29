# Use an official Go image
FROM golang:1.26@sha256:32c0e6e5c4f6707717051091b4d0b077464a679eaab563e11474efc5328e2aa5

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
