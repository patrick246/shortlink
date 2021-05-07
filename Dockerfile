FROM golang:1.16 as builder
WORKDIR /app
ENV CGO_ENABLED=0
COPY go.mod go.mod
COPY go.sum go.sum
RUN go mod download

COPY . .

RUN mkdir ./bin && go build -o ./bin -ldflags="-extldflags=-static" -tags osusergo,netgo -trimpath ./cmd/...

FROM gcr.io/distroless/static
LABEL org.opencontainers.image.source=https://github.com/patrick246/shortlink
LABEL org.opencontainers.image.authors=patrick246
LABEL org.opencontainers.image.licenses=AGPL-3.0
COPY --from=builder /app/bin/shortlink /shortlink
ENTRYPOINT ["/shortlink"]