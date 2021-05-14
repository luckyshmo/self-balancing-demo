build:
	go build
run:
	./sb-back -port=8080 -redirect=http://localhost:8081/ &
	./sb-back -port=8081 -redirect=http://localhost:8082/ &
	./sb-back -port=8082 -redirect=http://localhost:8080/ &
kill:
	pkill sb-back