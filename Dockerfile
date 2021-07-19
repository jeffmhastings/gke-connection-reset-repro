FROM golang:1.16 as build

WORKDIR /app
COPY go.* .
RUN go mod download
COPY cmd cmd
ARG CGO_ENABLED=0
RUN go build ./cmd/...

FROM scratch
COPY --from=build /app/gke-connection-reset-repro /gke-connection-reset-repro
CMD ["/gke-connection-reset-repro", "server"]
