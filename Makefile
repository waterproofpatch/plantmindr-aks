deploy_azure:
	(cd deployment/azure && bash deploy.sh)

deploy_k8s:
	(cd deployment/k8s && bash deploy.sh)

# builds docker images and pushes them to the acr, then restarts the k8s service
update_k8s: deploy_acr
	(cd deployment/k8s && bash update.sh)

deploy_acr:
	(cd deployment/acr && bash deploy.sh)

update: update_k8s

deploy: deploy_azure deploy_k8s

dev:
	docker-compose --file docker/docker-compose.yml up --build

devbox:
	docker build . -t plantmindr-devbox

run_devbox:
	docker run -v `pwd`:/workspace -it -w /workspace plantmindr-devbox /bin/bash