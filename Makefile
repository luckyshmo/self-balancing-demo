build:
	go build
run:
	./sb-back -port=8080 -redirect=http://localhost:8081/ -rn=1 &
	./sb-back -port=8081 -redirect=http://localhost:8082/ -rn=2 &
	./sb-back -port=8082 -redirect=http://localhost:8080/ -rn=3 &
kill:
	pkill sb-back
curl:
	curl http://localhost:8080/[1-7]