run-prod:
	docker build -t exponent-image .
	docker-compose up -d

stop-prod:
	docker-compose down