# Maud

<a href="https://www.youtube.com/watch?v=hWKB1Zxg84s"><img src="https://fc07.deviantart.net/fs71/f/2014/074/7/8/maud_pie_by_aa100500-d7aaxps.png" /></a>

**WIP** : This branch is currently going a lot of work.

## Local development

### Run on kubernetes/minikube

You will need:

- Skaffold

Run with skaffold with live-reload: `skaffold dev`

### Run on non-k8s docker

You will need:

- Docker compose

Run `docker-compose up`.

### Run locally

You will need:

- Go 1.11+
- Node.js
- MongoDB, running locally with no auth

Compile maud with

```sh
go install ./maud
```

Compile static assets with

```sh
npm i
npm start
```

Run maud

```sh
maud
```

Defaults to localhost/maud (db/collection). Config instructions coming soon(tm)
