# Use an official Go image
FROM golang:1.26.10@sha256:672c944cfd07599b5f1bb179869270cdc94c6254e41db4d431259693cb286b82

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
