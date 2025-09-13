# Plantmindr

## Prerequisites

``` bash
make run_devbox
```

## Explore

```bash
make init-k8s
# get the IP for the frontend:
echo $(kubectl get ingress static-website-frontend --output custom-columns='IP:.status.loadBalancer.ingress[0].ip')
```

## Deploy

```bash
az login
make deploy
```

## Testing

Dev k8s cluster

``` bash
go install sigs.k8s.io/kind@v0.23.0 && kind create cluster
```

``` bash
make dev
```

## Misc

* ingress controller setup: <https://spacelift.io/blog/kubernetes-ingress>
* To import secrets:

```bash
make run-devbox
[ ! -f oldplants/secret.env ] || export $(grep -v '^#' oldplants/secret.env | xargs)
```