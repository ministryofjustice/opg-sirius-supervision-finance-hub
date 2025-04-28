# Custom localstack image to allow for HTTPS (needed for S3 streaming)
FROM localstack/localstack:latest

COPY docker/localstack/localstack-cert.pem /usr/local/share/ca-certificates/localstack-cert.crt

RUN update-ca-certificates

EXPOSE 4566