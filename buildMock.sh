#!/bin/bash

if [[ -z "${GOPATH}" ]]; then
MOCKGEN=mockgen
else
MOCKGEN=$GOPATH/bin/mockgen
fi
mkdir -p mocks/cmd/web/
echo "Using $MOCKGEN"
$MOCKGEN -source cmd/web/menuCache.go -destination mocks/cmd/web/menuCache.go -package menuCacheMock -self_package github.com/alyrot/uksh-menu-parser/mocks/cmd/web

mkdir -p mocks/pkg/parser
$MOCKGEN -source pkg/parser/parser.go -destination mocks/pkg/parser/parser.go -package parserMock -self_package github.com/alyrot/uksh-menu-parser/mocks/pkg/parser
echo "Done generating mocks"