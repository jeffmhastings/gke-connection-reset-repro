# Debugging Network issues on GKE

The purpose of this project is to demonstrate issues affecting long lived TCP
connections on GKE when intranode visibility and network policies are enabled.
These issues appear to only affect network traffic between pods that are
located on the same node.

In order to run this example, we will be using a project in GCP to deploy a
GKE cluster with intranode visibility and network policies enabled. We will
then build and push a simple client/server container image to GCR and then
deploy the application on GKE.

Three pods will run when the application is deployed: a server and two clients.
Affinity rules will be used to ensure that One of the clients runs on the same
node as the server and the other will run on a different node.

By tailing the logs of the client application that is running on the same node
as the server (dubbed co-located-client), you will see connection
reset errors being logged periodically. The client running on the other node
(referred to as remote-client) will not be affected by the connection
resets, and you should not see any of these errors in it's logs.


## Assumptions

This README is written with the assumption that you have the command line tool
`gcloud` installed and configured to work with your GCP account. It also
assumes that you have a docker agent configured that can authenticate with
gcr.io for your project. The GCP cloud shell can be used for this example.

## 1. Deploy GKE

Deploy a standard GKE cluster with at least two nodes and with network policy
and intranode visibility enabled.

```bash
gcloud container clusters create gke-connection-reset-repro \
    --num-nodes=2 \
    --enable-intra-node-visibility \
    --enable-network-policy
```

## 2. Build and publish the container (Optional)

If you wish to build the container image and publish it to your own registry,
then you can do so.

Build the docker image and publish to your registry. Replace the value
`your-image-repository` with the repository you would like to use for
publishing the image.

```bash
docker build --tag your-image-repository .
docker push your-image-repository
```

Update [manifests/kustomization.yaml](./manifests/kustomization.yaml) to
uncomment the `images` section, and set `newName` to your image repository.

```yaml
images:
- name: jeffmhastings/gke-connecction-reset-repro
  newName: your-image-repository
```

## 3. Deploy the application to GKE

Deploy the application with `kubectl`

```bash
kubectl apply -k manifests
```

## 4. Monitor the application for errors

Check the logs for the server

```
kubectl logs -l app=netdebug,component=server
```

Check the logs for the client using an anti-affinity rule to run on a node
that the server pod is not running on.

```bash
kubectl logs -l app=netdebug,component=remote-client
```

Next, check the logs for the client using an affinity rule to run on the same
node as the server.

```bash
kubectl logs -l app=netdebug,component=co-located-client
```

In the pod logs for the affinity client, you should see error messages like
the following. It may take some time for the error to appear, but usually less
than a few minutes.

```
2021/07/19 20:39:38 read tcp 10.44.0.33:33326->10.48.14.50:8080: read: connection reset by peer
```

After that, the client will reconnect and continue making requests, and continue
to be affected by unexpected connection reset errors.
