#
# 1. Build Container
#
FROM golang:1.12.1 AS build

ENV GOOS=linux \
    GOARCH=amd64

RUN mkdir -p /src

# First add modules list to better utilize caching
COPY go.sum go.mod /src/

WORKDIR /src

# Download dependencies
RUN go mod download

COPY . /src

# Build components.
# Put built binaries and runtime resources in /app dir ready to be copied over or used.
RUN CGO_ENABLED=0 go install -ldflags="-w -s" && \
    mkdir -p /app && \
    cp -r "$GOPATH/bin/tiles" /app/

#
# 2. Runtime Container
#
FROM alpine:3.9

ENV TZ=UTC \
    PATH="/app:${PATH}"

RUN apk add --no-cache \
      tzdata \
      ca-certificates \
      bash \
    && \
    cp --remove-destination /usr/share/zoneinfo/${TZ} /etc/localtime && \
    echo "${TZ}" > /etc/timezone && \
    mkdir -p /var/log && \
    chgrp -R 0 /var/log && \
    chmod -R g=u /var/log

WORKDIR /app

# expose default port of the application
EXPOSE 1372

RUN touch /app/shards.yaml
COPY --from=build /app /app/

CMD ["./tiles"]
