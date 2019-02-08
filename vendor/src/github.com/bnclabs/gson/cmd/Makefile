build:
	go build -o gson
	go build -tags n1ql -o gsonn1ql

test: build
	go test -race -test.run=. -test.bench=. -test.benchmem=true
	./sort.sh
	./transforms.sh

clean:
	rm -rf gson gsonn1ql cmd
