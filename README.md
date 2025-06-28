# go-jq-server
Takes a jq query over HTTP, passes it to the jq utility and returns results over HTTP

Run the dockerfile on a port of your choosing:
docker run -p 8080:8080 jq_server

And then try it out!

## Request
curl -X "POST" "http://localhost:8080/?filter=length" \
     -H 'Content-Type: Content-Type: application/json' \
     -d "{set your body string}"
