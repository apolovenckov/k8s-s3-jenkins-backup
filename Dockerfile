############################
# Builder: compile Go binary
############################
FROM golang:1.21-alpine AS builder
WORKDIR /src
COPY . .
RUN CGO_ENABLED=0 go build -o /out/backup ./cmd/backup

############################
# Runtime: tools + our binary
############################
FROM alpine:3.6

# Install dependencies (awscli + helpers)
RUN apk -v --update add \
        python \
        py-pip \
        groff \
        less \
        mailcap \
        curl \
        && \
    pip install --upgrade awscli s3cmd python-magic && \
    apk -v --purge del py-pip && \
    rm /var/cache/apk/*

# Install kubectl
RUN curl -LO "https://storage.googleapis.com/kubernetes-release/release/$(curl -s https://storage.googleapis.com/kubernetes-release/release/stable.txt)/bin/linux/amd64/kubectl" && \
    chmod +x ./kubectl && \
    mv ./kubectl /usr/local/bin/kubectl

# No baked-in app defaults; configure via env at runtime

# Copy Go binary
COPY --from=builder /out/backup /usr/local/bin/backup

# Run backup tool
ENTRYPOINT ["/usr/local/bin/backup"]
