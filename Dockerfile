# this is for the devbox
FROM mcr.microsoft.com/azure-cli:2.63.0
RUN apk update & apk add make
RUN az bicep upgrade
RUN az aks install-cli
RUN wget -O jq https://github.com/stedolan/jq/releases/download/jq-1.6/jq-linux64
RUN chmod +x ./jq
RUN mv jq /usr/local/bin