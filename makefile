docker-run:
	@docker rmi goim/gateway
	@docker rmi goim/service
	@docker build -t goim/gateway -f ./deployment/gateway/Dockerfile .
	@docker build -t goim/service -f ./deployment/service/Dockerfile .
	@docker-compose -f docker/docker-compose.yml up -d
