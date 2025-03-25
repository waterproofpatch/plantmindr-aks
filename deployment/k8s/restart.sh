#!/bin/bash
kubectl rollout restart deployment static-website-backend
kubectl rollout restart deployment static-website-frontend