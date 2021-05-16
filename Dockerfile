FROM golang:1.16 AS builder

WORKDIR /workspace
COPY go.mod go.sum *.go /workspace/

RUN go build -o /bin/server .

FROM gcr.io/distroless/base-debian10

COPY --from=builder /bin/server /bin/

ENTRYPOINT ["/bin/server"]
