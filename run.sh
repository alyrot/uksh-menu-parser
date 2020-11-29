#!/bin/bash

sudo docker run -p 8080:80 -e  SERVER_LISTEN=:80 uksh-menu-api
