dataGenerator:
	mkdir -p bin
	go build -o bin/dataGenerator mainGenerateData.go

candleApi:
	mkdir -p bin
	go build -o bin/candle-api main.go
