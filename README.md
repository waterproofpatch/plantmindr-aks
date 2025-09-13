# Plantmindr

## Prerequisites

Make sure `azure-cli` is installed:

``` bash
brew install azure-cli
```

``` bash
make run_devbox
```

## Explore

```bash
make init-k8s
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