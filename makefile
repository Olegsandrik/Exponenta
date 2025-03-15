run-prod:
	docker build -t exponent-image .
	docker-compose -f Docker-compose.yml up -d

stop-prod:
	docker-compose -f Docker-compose.yml down