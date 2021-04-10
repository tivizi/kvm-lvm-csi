	
all: node controller

node: 
	go build -o bin/ ./cmd/klc-node
controller: 
	go build -o bin/ ./cmd/klc-controller
clean:
	rm bin/ -rf