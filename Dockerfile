FROM golang:1.25-alpine AS builder

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .

ARG VERSION=dev
ARG COMMIT=none
ARG BUILD_DATE=unknown

RUN CGO_ENABLED=0 go build \
    -ldflags "-s -w \
      -X github.com/Ommanimesh2/rift/cmd.version=${VERSION} \
      -X github.com/Ommanimesh2/rift/cmd.commitHash=${COMMIT} \
      -X github.com/Ommanimesh2/rift/cmd.buildDate=${BUILD_DATE}" \
    -o /rift .

FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=builder /rift /usr/local/bin/rift

ENTRYPOINT ["rift"]
