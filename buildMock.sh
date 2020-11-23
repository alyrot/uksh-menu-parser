#!/bin/bash

mkdir -p mocks/cmd/web/
$GOPATH/bin/mockgen -source cmd/web/menuCache.go -destination mocks/cmd/web/menuCache.go -package menuCacheMock

mkdir -p mocks/pkg/parser
$GOPATH/bin/mockgen -source pkg/parser/parser.go -destination mocks/pkg/parser/parser.go -package parserMock
