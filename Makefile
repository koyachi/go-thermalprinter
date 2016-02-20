
_printertest:
	GOOS=linux GOARCH=arm go build -o bin/printertest -v github.com/koyachi/go-thermalprinter/printertest/printertest

helloworld:
	GOOS=linux GOARCH=arm go build -o bin/helloworld -v github.com/koyachi/go-thermalprinter/printertest/helloworld

atkinson:
	GOOS=linux GOARCH=arm go build -o bin/atkinson -v github.com/koyachi/go-thermalprinter/printertest/atkinson
