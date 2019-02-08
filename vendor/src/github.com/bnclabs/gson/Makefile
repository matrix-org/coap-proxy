build:
	go build
	make -C cmd build

test: build
	make -C cmd test
	go test -race -timeout 4000s -test.run=. -test.bench=xxx -test.benchmem=true
	go test -timeout 4000s -test.run=xxx -test.bench=. -test.benchmem=true

heaptest:
	go test -gcflags '-m -l' -timeout 4000s -test.run=. -test.bench=. -test.benchmem=true > escapel 2>&1
	grep "^\.\/.*escapes to heap" escapel | tee escapelines
	grep panic *.go | tee -a escapelines

coverage:
	go test -coverprofile=coverage.out
	go tool cover -html=coverage.out
	rm -rf coverage.out
