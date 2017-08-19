all: route-switcher

route-switcher: $(shell find . -name \*.go)
	@echo Starting route-switcher binary build
	CGO_ENABLED=0 go build -o route-switcher route-switcher.go
	@echo Finished route-switcher binary build