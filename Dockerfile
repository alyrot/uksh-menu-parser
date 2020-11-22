FROM ubuntu:rolling

#our app uses these cli utilities
RUN apt-get update && apt-get install -y tesseract-ocr tesseract-ocr-deu poppler-utils golang ca-certificates

ADD . /usr/local/go/src/github.com/alyrot/menuToText


#setup timezone
ENV TZ=Europe/Berlin
ENV ZONEINFO=/zoneinfo.zip

#app specific env
ENV SERVER_LISTEN=:80
ENV GOPATH=/usr/local/go

WORKDIR /usr/local/go/src/github.com/alyrot/menuToText
RUN ["go", "mod", "download"]
RUN ["go", "mod", "vendor"]
#RUN ["go", "get", "github.com/golang/mock/mockgen@v1.4.4"]
#RUN ["./buildMock.sh"]
RUN ["go", "test", "./..."]
WORKDIR /usr/local/go/src/github.com/alyrot/menuToText/cmd/web
RUN go build

ENTRYPOINT ["/usr/local/go/src/github.com/alyrot/menuToText/cmd/web/web"]

EXPOSE 80
EXPOSE 443
