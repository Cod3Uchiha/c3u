FROM golang:1.23-bookworm AS build
WORKDIR /src
COPY go.mod ./
COPY cmd ./cmd
COPY internal ./internal
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/c3u ./cmd/c3u

FROM gcr.io/distroless/base-debian12:nonroot
COPY --from=build /out/c3u /usr/local/bin/c3u
VOLUME ["/data"]
EXPOSE 39333
ENTRYPOINT ["/usr/local/bin/c3u"]
CMD ["node", "--network", "mainnet", "--data", "/data", "--listen", ":39333"]
