FROM node:14 AS node-builder

COPY frontend /frontend

RUN npm ci && npm run build

FROM golang:1.16 AS go-builder

WORKDIR /workspace
COPY go.mod go.sum *.go /workspace/

RUN go build -o /bin/server .

FROM gcr.io/distroless/base-debian10

COPY --from=go-builder /bin/server /bin/

ENTRYPOINT ["/bin/server"]
