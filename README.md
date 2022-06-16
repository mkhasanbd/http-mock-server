# http-mock-server



Commands to run service



Non HTTPS

    go run main.go -ip=<IP> -port=<PORT> -config=<CONFIG>.yaml -output=<OUTPUTFILE> -verbose

HTTPS

    go run main.go -ip=<IP> -port=<PORT> -https -cert=<server>.crt -key=<client-key>.key -config=<CONFIG>.yaml -output=<OUTPUTFILE> -verbose

example : 

go run main.go -ip=localhost -port=9999  -config=config.yaml -output=debug.log -verbose

go run main.go -port=9999 -https -cert=https-cert-and-keys/server.crt -key=https-cert-and-keys/server.key -config=config.yaml -output=debug.log -verbose
