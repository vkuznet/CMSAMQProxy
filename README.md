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
