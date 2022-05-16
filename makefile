dcup:
	docker-compose up --build -d

build:
	docker-compose exec ghen go build
	docker-compose exec ghen zip ghen.zip ghen

run:
	docker-compose exec ghen /go/src/ghen/ghen

bash:
	docker-compose exec ghen bash
	