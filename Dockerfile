FROM ubuntu:rolling
#setup timezone
ENV TZ=Europe/Berlin
ENV ZONEINFO=/zoneinfo.zip

#our app uses these cli utilities
RUN apt-get update && apt-get install -y tesseract-ocr tesseract-ocr-deu poppler-utils ca-certificates wget build-essential git

#install specific go version
RUN ["wget", "https://golang.org/dl/go1.15.5.linux-amd64.tar.gz"]
RUN ["tar", "-C", "/usr/local", "-xzf", "go1.15.5.linux-amd64.tar.gz"]
ENV GOPATH=/root/go/
ENV GOROOT=/usr/local/go
ENV GO111MODULE=on
ENV PATH="${PATH}:/usr/local/go/bin:$GOPATH/bin"



#build our app
ADD . /root/go/github.com/alyrot/uksh-menu-parser
WORKDIR /root/go/github.com/alyrot/uksh-menu-parser
RUN go build ./...
RUN go build ./cmd/web
RUN go test ./...


ENTRYPOINT ["/root/go/github.com/alyrot/uksh-menu-parser/web"]

