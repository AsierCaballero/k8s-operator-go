FROM golang:1.22-alpine AS builder

ARG TARGETOS=linux
ARG TARGETARCH=amd64

WORKDIR /workspace

RUN apk add --no-cache git ca-certificates

COPY go.mod go.mod
COPY go.sum go.sum
RUN go mod download

COPY cmd/ cmd/
COPY api/ api/
COPY controllers/ controllers/
COPY internal/ internal/

RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build -ldflags="-s -w" -o manager cmd/manager/main.go

FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata

RUN adduser -D -u 1000 appuser

COPY --from=builder /workspace/manager /manager

USER appuser

EXPOSE 8080 8081 9443

ENTRYPOINT ["/manager"]
