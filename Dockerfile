FROM golang:alpine AS builder

ARG VERSION

LABEL stage=gobuilder

ENV CGO_ENABLED 0

RUN apk update --no-cache && apk add --no-cache tzdata

WORKDIR /build

ADD go.mod .
ADD go.sum .
RUN go mod download
COPY . .
RUN go build -ldflags="-s -w" -o /app/hn main.go


FROM scratch

ARG VERSION

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=builder /usr/share/zoneinfo/Asia/Shanghai /usr/share/zoneinfo/Asia/Shanghai
ENV TZ Asia/Shanghai
ENV HN_VERSION $VERSION

WORKDIR /app
COPY --from=builder /app/hn /app/hn

ENV GIN_MODE release

CMD ["/app/hn"]