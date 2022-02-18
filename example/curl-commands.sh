#!/bin/sh -e

echo "----- Gets the default, which is English as JSON -----"
echo
echo "curl -i -H 'Accept:' http://localhost:8080/"
curl -i -H 'Accept:' http://localhost:8080/
echo

echo "----- Gets French as JSON -----"
echo
echo "curl -i -H 'Accept: application/json' -H 'Accept-Language: fr' http://localhost:8080/"
curl -i -H 'Accept: application/json' -H 'Accept-Language: fr' http://localhost:8080/
echo

echo "----- Gets English as JSON because there is no German and the first language offered is used instead -----"
echo
echo "curl -i -H 'Accept-Language: de' http://localhost:8080/"
curl -i -H 'Accept-Language: de' http://localhost:8080/
echo

echo "----- Gets French as HTML using the page _index.html -----"
echo
echo "curl -i -H 'Accept: text/html' -H 'Accept-Language: fr' http://localhost:8080/"
curl -i -H 'Accept: text/html' -H 'Accept-Language: fr' http://localhost:8080/
echo

echo "----- Gets Russian as HTML using the page home.html -----"
echo
echo "curl -i -H 'Accept: application/xhtml+xml' -H 'Accept-Language: ru' http://localhost:8080/home.html"
curl -i -H 'Accept: application/xhtml+xml' -H 'Accept-Language: ru' http://localhost:8080/home.html
echo

# curl -i -H 'Accept: application/xhtml+xml' -H 'Accept-Language: ru' http://localhost:8080/home.html
