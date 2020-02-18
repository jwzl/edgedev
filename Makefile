.PHONY: edgedev

edgedev:
	@export GO111MODULE=on && \
	export GOPROXY=https://goproxy.io && \
	go build edgedev.go
	
	@chmod 777 edgedev


.PHONY: clean
clean:
	@rm -rf edgedev
	@echo "[clean Done]"
