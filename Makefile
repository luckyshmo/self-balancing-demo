build:
	go build
run: #номер порта приложения/на какой бэкенд перенаправлять трафик/колчиество запросов, которое бэкенд может обработать за секунду
	./sb-back -port=8080 -redirect=http://localhost:8081/ -rn=1 &
	./sb-back -port=8081 -redirect=http://localhost:8082/ -rn=2 &
	./sb-back -port=8082 -redirect=http://localhost:8080/ -rn=3 &
kill:
	pkill sb-back
curl:
	curl http://localhost:8080/[1-7]