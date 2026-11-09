FROM golang:1.26.3-bookworm AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -o /out/ ./cmd/...

FROM debian:bookworm-slim
RUN useradd --system --uid 10001 app
WORKDIR /app
COPY --from=build /out/ /app/
USER app
CMD ["/app/gateway"]
