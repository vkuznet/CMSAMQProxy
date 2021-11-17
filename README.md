### CMSAMQProxy
Proxy server for CMS AMQ brokers. Build code using `make`.
To run, please create new config file like this one:
```
cat > config.json << EOF
{
    "base": "/cmsamqproxy",
    "port": 8300,
    "logFile": "/tmp/cmsamqproxy.log",
    "stompURI": "uri",
    "stompLogin": "login",
    "stompPassword": "password",
    "endpoint": "endpoint",
    "producer": "producer-name",
    "verbose": 0
}
EOF
```
and run the code as following
```
cmsamqproxy -config config.json
```

A client can send request to our server as following:
```
# get server status
curl -v http://localhost:8300/status

# use JSON payload with list of records
curl -X POST -H "Content-type: application/json" -d '[{"foo":1},{"bla":1}]' http://localhost:8300/cmsamqproxy
[{"ids":["91c9446e007b46a6991e2494acc218f6","22344dae3d8b45c8859e4319dd2097d7"],"status":"ok"}]


# use gzip encoding with JSON payload
curl -X POST -H "Content-type: application/json" -H "Content-Encoding: gzip" --data-binary @/tmp/data.json.gz http://localhost:8300/cmsamqproxy
[{"ids":["5707ac0208de4477b3698250ab1dd08d","f031345cf6f84a349c57c73842517f45"],"status":"ok"}]
```
