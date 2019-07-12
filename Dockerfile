FROM golang:1.12.7-alpine3.10 AS BUILDER

WORKDIR /go/src/

ENV GO111MODULE on
ENV GOPROXY https://goproxy.io

ADD . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo ./example/main.go

CMD [ "./main" ]

FROM alpine:3.9.2

WORKDIR /go/src/

COPY --from=BUILDER /go/src/main main

CMD [ "./main" ]