# Plantmindr

## Prerequisites

Make sure `azure-cli` is installed:

``` bash
brew install azure-cli
```

``` bash
make devbox
make run_devbox
```

## Explore

```bash
make init-k8s
```

## Deploy

```
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

## DB Utils

Make sure your local DB is up (`make dev`) or else you might get error about app-db-user not existing.

``` bash
cd docker/backend/tools/DatabaseCopyProject
dotnet build
source ../../../../secret.env
./bin/Debug/net8.0/DatabaseCopyProject
```

## Misc

* ingress controller setup: <https://spacelift.io/blog/kubernetes-ingress>
