
printertest:
	GOOS=linux GOARCH=arm go build -o bin/printertest -v github.com/koyachi/go-thermalprinter/examples/printertest

helloworld:
	GOOS=linux GOARCH=arm go build -o bin/helloworld -v github.com/koyachi/go-thermalprinter/examples/helloworld

atkinson:
	GOOS=linux GOARCH=arm go build -o bin/atkinson -v github.com/koyachi/go-thermalprinter/examples/atkinson
