# expconf

Trying to make it a config store service with great features:
1. App/Env realized and seperated
2. Quick response and realtime config
3. Preconfigurable config store and can be configured with default values
4. More QPS(current 1000+qps read) I want it to be 1M
5. Version controllable, with fallback function at both key/group level
6. Passive update of config data at client side

### Usage
1. start mysql server at local (should be configured as root and with no password)
2. go build main.go
3. ./main
