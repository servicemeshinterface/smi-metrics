FROM golang:1.12-alpine as builder
ARG NETRC

WORKDIR /src

RUN apk add --no-cache \
    git

COPY go.mod .
COPY go.sum .

RUN echo "${NETRC}" > ~/.netrc
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build ./cmd/smi-metrics/...

FROM scratch

WORKDIR /src

COPY --from=builder /src/smi-metrics /

CMD /smi-metrics
