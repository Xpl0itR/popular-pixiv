FROM golang:1.15-alpine

COPY . /go/src/popular-pixiv
WORKDIR /go/src/popular-pixiv

RUN go get
RUN go build

CMD popular-pixiv --address :80 --refresh_token $REFRESH_TOKEN

EXPOSE 80