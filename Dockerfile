FROM golang:1.12.1-alpine3.9 AS BUILDER

WORKDIR /go/src/

ENV GO111MODULE on

ADD . .

RUN CGO_ENABLED=0 GOOS=linux go build -mod=vendor -a -installsuffix cgo ./example/main.go

CMD [ "./main" ]

FROM alpine:3.9.2

WORKDIR /go/src/

COPY --from=BUILDER main main

CMD [ "./main" ]