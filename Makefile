build:
	go build -o assetgen assetsgen/main/asssetsgen.go
	./assetgen
	go build -buildmode=plugin -o plugins/vmwarevim.so plugins/vmwarevim.go plugins/vmwarensx.go plugins/vmwareconfig.go plugins/vmwarediscovery.go
	go build -o jettison main.go
