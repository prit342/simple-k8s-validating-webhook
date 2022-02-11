# simple-k8s-vlidating-webhook

This repo contains a validating kubernetes admission webhook written in Go.  It reads the following enviroment variables:

- CERT_PATH - default value is set to "/source/cert.pem". This is the certificate to server the TLS traffic
- KEY_PATH" - default value is set "/source/key.pem". This is the private Key of the TLS certificate
- PORT - default valie is set to 3000. Port where the validating web-hook will listen
- ANNOTATION - Default value is set to "example.com/validate". The default annotation to check on the namespace. If the value of this annotiation is to true then only the object is validated else the validation is skipped
- LABEL - Default value is set to "owner". This is the label on the Pod object that the webhook controlled will check for and if it is present then only the object will be allowed to be created.

## Flow

- The validation webhook is triggered for a Pod CREATE operation
- The webhook checks if the namespace where the object is created has the correct annotation set. This annotation is defined by the environment variable `ANNOTATION`. The default value of this is set to `example.com/validate`.
- If the namespace has the annotation `example.com/validate` and if the value of that annotation is set to `true` then the webhook will check if the label defined by the environment variable `LABEL` is present on the object.If the annotation is not present or is set to false then the validation is skipped and the reason is logged



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
docker build . --tag=webhook-server:1.5 --no-cache
kind load docker-image webhook-server:1.5 --name test
```

- Since, the webhook needs to check the annotations on the namespaces, it needs a service account with permission to read a namespace

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

- Create a namespace and annotate it with `example.com/validate:true`

```bash
kubectl create ns test-ns
kubectl annotate ns test-ns example.com/validate=true
```

- if you apply the Pod.yaml that does not have the label `owner` in the test-ns, the request is rejected by the validating admission controller.

```bash
kubectl apply -f k8s-manifests/pod.yaml -n test-ns                                                                   

Error from server: error when creating "k8s-manifests/pod.yaml": admission webhook "webhook-server.webhook-demo.svc" denied the request: Denied because the Pod is missing label owner
```

- If you edit `k8s-manifests/pod.yaml` the add owner label to the Pod, the request will not be blocked and the Pod will be created

- If you remove the annotation from the namespace or set it to `example.com/validate:false`, you will still be able to create the resource without the label `owner`

