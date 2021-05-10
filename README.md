# simple-k8s-vlidating-webhook

This repo contains a validating kubernetes admission controller written in golang. The validation webhook is triggered for a Pod CREATE operation and will reject the operation if the label `owner` is missing from the Pod manifest. 

## Installation

I am documenting the steps with [`kind`](https://kind.sigs.k8s.io/docs/user/quick-start/). You can use any K8s cluser.

- Create a K8s cluster using kind, make sure that you enable `ValidatingAdmissionWebhook` plugin. A sample kind cluster manifest has been provided in the k8s-manifests repo.

```bash
kind create cluster --name test --config ./k8s-manifests/kind-cluster.yaml --image kindest/node:v1.20.2
```

- Generate self signed certs for quick testing (obviously this is a big NO in production environments). The webhook by default reads the cert and key from `/source/cert.pem` and `/source/key.pem` respectively. We will create a kubernetes secret that will later be mounted inside the webhook pod.

```bash
cd certs
./generate-certs.sh
kubectl create ns webhook-demo
kubectl create secret generic webhook-certs --from-file=key.pem=webhook-server-tls.key --from-file=cert.pem=webhook-server-tls.crt -n webhook-demo
```

- Build the image using the docker file and load it in the `kind` cluster:

```bash
cd simple-validating-webhook
docker build . --tag=webhook-server:1.5 --no-cache && kind load docker-image webhook-server:1.5 --name test
```

- Deploy the webhook to the K8s cluster and wait for the Pod to become healthy:

```bash
kubectl apply -f k8s-manifests/webhook-deployment-service.yaml
```

- In order for K8s API Server to validate the HTTPS service endpoint of the webhook service, we need to add the CA Public certificate  generated in the previous step into to the ValidatingWebhookConfiguration object.

```bash

# you should be in the root of the repo
WEBHOOKCA=$(cat certs/ca.crt | openssl base64 -A)
sed -i.bak "s/CHANGE_THIS_CA/$WEBHOOKCA/g" k8s-manifests/ValidatingWebhookConfiguration.yaml
kubectl apply -f k8s-manifests/ValidatingWebhookConfiguration.yaml
kubectl get validatingwebhookconfigurations
```

## Testing

- Apply the Pod.yaml that does not have the label `owner` added to to it and then you will see the request is rejected by our validating admission controller.

```bash
kubectl apply -f k8s-manifests/pod.yaml                                                                           130 ↵
```

```bash
Error from server: error when creating "k8s-manifests/pod.yaml": admission webhook "webhook-server.webhook-demo.svc" denied the request: Denied because the Pod is missing label owner
```

- If you add owner label to the Pod, the request will not be blocked and the Pod will be created
