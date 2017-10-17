FROM golang:1.8
WORKDIR /go/src/app
COPY . .
RUN go-wrapper download   # "go get -d -v ./..."
RUN go-wrapper install    # "go install -v ./..."
EXPOSE 9143
RUN mkdir /data
CMD ["go-wrapper", "run"] # ["app"]
