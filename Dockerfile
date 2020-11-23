FROM ubuntu:rolling
#setup timezone
ENV TZ=Europe/Berlin
ENV ZONEINFO=/zoneinfo.zip

#go specific env



#app specific env
ENV SERVER_LISTEN=:80

#our app uses these cli utilities
RUN apt-get update && apt-get install -y tesseract-ocr tesseract-ocr-deu poppler-utils ca-certificates wget build-essential

#install specific go version
RUN ["wget", "https://golang.org/dl/go1.15.5.linux-amd64.tar.gz"]
RUN ["tar", "-C", "/usr/local", "-xzf", "go1.15.5.linux-amd64.tar.gz"]
ENV PATH="${PATH}:/usr/local/go/bin"
ENV GOPATH=/root/go/



#build our app
ADD . /usr/local/go/src/github.com/alyrot/menuToText
WORKDIR /usr/local/go/src/github.com/alyrot/menuToText
RUN ["go", "mod", "download"]
RUN ["go", "mod", "vendor"]
RUN ["go", "get", "github.com/golang/mock/mockgen@v1.4.4"]
RUN ["./buildMock.sh"]
RUN ["go", "test", "./..."]
WORKDIR /usr/local/go/src/github.com/alyrot/menuToText/cmd/web
RUN go build

ENTRYPOINT ["/usr/local/go/src/github.com/alyrot/menuToText/cmd/web/web"]

EXPOSE 80
EXPOSE 443
