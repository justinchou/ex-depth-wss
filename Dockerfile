FROM golang AS build-env
ENV GO111MODULE=on
WORKDIR /go/src/app
ADD . /go/src/app
RUN go mod tidy
RUN GOOS=linux GOARCH=386 go build -v -o /go/src/app/ex-depth

FROM alpine
WORKDIR /data
RUN apk add -U tzdata
RUN ln -sf /usr/share/zoneinfo/Asia/Shanghai  /etc/localtime
COPY --from=build-env /go/src/app/ex-depth /data/ex-depth
ENTRYPOINT [ "./ex-depth" ]
