# Stage 1: Build the custom Go runtime
FROM debian:bookworm-slim AS goruntime-builder

RUN apt-get update && apt-get install -y bash build-essential curl git

# Install bootstrap Go compiler for building custom runtime
ENV GOROOT_BOOTSTRAP=/usr/local/go

RUN curl -fsSL https://go.dev/dl/go1.27.3.linux-amd64.tar.gz | tar -C /usr/local -xzf -

COPY goPatch /Advocate/goPatch
WORKDIR /Advocate/goPatch/src
RUN bash make.bash

# Stage 2: Build the Go app using the standard Go runtime
FROM golang:1.25 AS app-builder

WORKDIR /ADVOCATE/advocate
COPY advocate /ADVOCATE/advocate
COPY goPatch /ADVOCATE/goPatch

RUN go build -o app .

# Stage 3: Final runtime image
FROM debian:bookworm-slim

WORKDIR /Advocate

COPY --from=goruntime-builder /Advocate/goPatch ./goPatch
COPY --from=app-builder /Advocate/advocate/app ./advocate/app

WORKDIR /prog

WORKDIR /Advocate/advocate

ENTRYPOINT ["./app"]
