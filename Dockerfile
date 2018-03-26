# build stage
FROM golang:alpine AS build-env

ADD . /go/src/github.com/wenwu449/vsts-pr-review-sample

WORKDIR /go/src/github.com/wenwu449/vsts-pr-review-sample/
RUN go test -v && CGO_ENABLED=0 GGOS=linux go build -o pr-review

# final stage
FROM alpine

RUN apk add --no-cache ca-certificates
WORKDIR /pr-review
COPY --from=build-env /go/src/github.com/wenwu449/vsts-pr-review-sample/pr-review .

CMD ["./pr-review"]