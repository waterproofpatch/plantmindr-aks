az_login:
	az login --use-device-code --tenant 9d2c6922-fbb0-401b-a991-ca9b417eefec

deploy_azure:
	(cd deployment/azure && bash deploy.sh)

deploy_k8s:
	(cd deployment/k8s && bash deploy.sh)

# builds docker images and pushes them to the acr, then restarts the k8s service
update_k8s: update_images restart_k8s

restart_k8s:
	(cd deployment/k8s && bash restart.sh)

update_images: update_frontend update_backend

update_backend:
	(cd deployment/k8s && bash update_images.sh backend)
	(cd deployment/k8s && bash restart.sh)

update_frontend:
	(cd deployment/k8s && bash update_images.sh frontend)
	(cd deployment/k8s && bash restart.sh)

update: update_k8s

deploy: deploy_azure deploy_k8s

dev:
	docker-compose --file docker/docker-compose.yml up --build

init-k8s: az_login
	(cd deployment/k8s && bash init.sh)

devbox:
	docker build . -t plantmindr-devbox

run_devbox: devbox
	docker run -v `pwd`:/workspace -it -w /workspace plantmindr-devbox /bin/bash

backend_logs:
	kubectl logs -f -l app=static-website-backend --tail=1000

.PHONY: get_ip
get_ip:
	@kubectl get ingress static-website-frontend --output custom-columns='IP:.status.loadBalancer.ingress[0].ip'