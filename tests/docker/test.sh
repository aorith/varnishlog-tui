#!/usr/bin/env bash

curl 127.0.0.1:8080/esi/ -H'Host: www.example1.com'

# same connection session
curl -H 'Host: www.example2.com' 127.0.0.1:8080/esi2/ 127.0.0.1:8080/esi2/
sleep 0.5

curl 127.0.0.1:8080/where-is-it/ -H'Host: other.example1.com' -v
sleep 0.5
curl 127.0.0.1:8080/admin -H'Host: bad.example1.com' -v
