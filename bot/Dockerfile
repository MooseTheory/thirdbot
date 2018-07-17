FROM golang:1.10.3 as builder
WORKDIR /go/src/app
COPY ./*.go /go/src/app/
COPY vendor/ /go/src/app/vendor/
RUN CGO_ENABLED=0 GOOS=linux go install -v ./...

FROM alpine:latest
RUN apk --no-cache add tzdata ca-certificates
COPY --from=builder /go/bin/app /
CMD ["/app"]