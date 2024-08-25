# Stuff

## Prerequisites

Make sure `azure-cli` is installed:

``` bash
brew install azure-cli
```

``` bash
make install
```

## Deploy

``` bash
(cd deployment && bash deploy.sh)
```

## Using Docker for azure-cli

``` bash
make devbox
make run_devbox
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

``` bash
cd docker/backend/tools
dotnet build
source ../../../../secret.env
./bin/Debug/net8.0/DatabaseCopyProject
```

## Misc

* ingress controller setup: <https://spacelift.io/blog/kubernetes-ingress>
