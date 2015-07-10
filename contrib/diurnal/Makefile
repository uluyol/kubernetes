.PHONY: build push vet test clean

TAG = 0.4.1
REPO = uluyol/kube-diurnal

BIN = dc

dc: dc.go
	CGO_ENABLED=0 godep go build -a -installsuffix cgo dc.go

vet:
	godep go vet .

test:
	godep go test .

build: $(BIN)
	docker build -t $(REPO):$(TAG) .

push:
	docker push $(REPO):$(TAG)

clean:
	rm -f $(BIN)
