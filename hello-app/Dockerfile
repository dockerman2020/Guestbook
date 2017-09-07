FROM golang:1.8-alpine
ADD . /hello-app
WORKDIR /hello-app
RUN go build -o main .
ENV PORT 8080
CMD ["/hello-app/main"]
