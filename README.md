This is for exploring GoogleCloudPlatform agones.

More info at https://github.com/GoogleCloudPlatform/agones.

This example is a simple Ping/Pong websocket example.
The websocket could be used to stream game data to a JS game client.

GKE
===
Start a new cluster. I just use the defaults. Run the `gcloud` command to add the cluster to your kubeconfig.

```
kubectl create clusterrolebinding cluster-admin-binding \
  --clusterrole cluster-admin --user `gcloud config get-value account`
```

```
 gcloud compute firewall-rules create game-server-firewall \
  --allow tcp:7000-8000 \
  --description "Firewall to allow game server udp traffic"
```

Install agones
==============

`kubectl apply -f https://github.com/GoogleCloudPlatform/agones/raw/release-0.2.0/install/yaml/install.yaml`

Building
========

Find and replace `YOUR_PROJECT_ID` in the Go and YAML files with your `gcr.io` repository project ID. I used the same project that my GKE cluster was created in to make accessing the image repository easy.

`sed -i 's/YOUR_PROJECT_ID/example-123456/g' *`
`make bulid`
`make build-image`

Push Images
===========

`make push-image`

Deploying
=========

`kubectl apply -f yaml/launcher`

`kubectl get -n simple-launcher service`

Testing
=======
Now visit the simple-launcher service. You should see a link to create a new gameserver. After clicking the link, a new gameserver will be created.

The details for the gameserver will load and you will get another link to access the new gameserver instance.

The gameserver instance is a VueJs/WebSocket single page that has a Ping/Pong app that talks to the gameserver. It also has a button for stoppind the gameserver.
