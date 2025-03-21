# this is for the devbox
FROM mcr.microsoft.com/azure-cli:2.64.0
RUN tdnf install make wget -y

RUN az bicep upgrade
RUN az aks install-cli

RUN wget -O jq https://github.com/stedolan/jq/releases/download/jq-1.6/jq-linux64
RUN chmod +x ./jq
RUN mv jq /usr/local/bin