# Graphviz Renderer - Web

## What is this?
This is a dockerized replacement for the [GraphViz](https://graphviz.org/) functionality of the
[Google Image Charts API](https://developers.google.com/chart/image/), which I've been using as
a backend for [graphs.grevian.org](graphs.grevian.org) up until April of 2020 despite it being
marked deprecated in 2012, and being "turned off" in 2019

## How does it work?

Run the server or deploy the docker container and make `POST` requests to `service:8080/chart` with a 
`Content-Type` header of `application/x-www-form-urlencoded` and a body containing the following 3 parameters
* `chof` Output format, must be `png` for now
* `cht` Chart type, can be one of circo, dot, fdp, neato, nop, nop1, nop2, osage, patchwork, sfdp, twopi
   * See the [Layout Manual Pages](https://www.graphviz.org/documentation/) in the graphviz documentation for details
* `chl` A [dot formatted](https://en.wikipedia.org/wiki/DOT_(graph_description_language)) graph description to be
   rendered, [click here to see some examples](https://graphs.grevian.org/example)

The response will be a simple png image with the rendered graph and a `200` status, or a `400` or `500` status with a
 plaintext error message

## Deployment
This service can be built as a local go service, or using the included Dockerfile and deployed wherever you want, 
it listens on port 8080 by default and if you set the PORT environment variable it'll listen on that port.

Build the container using [google cloud build](https://cloud.google.com/cloud-build), and push it to a project registry
```
gcloud builds submit --tag gcr.io/[PROJECT_ID]/gv-renderer
```

Deploy the container via [Google Cloud Run](https://cloud.google.com/run)
```
gcloud run deploy --image gcr.io/[PROJECT_ID]/gv-renderer --platform managed --memory=64 --max-instances=5
```

## Contact

Josh Hayes-Sheen, [grevian@gmail.com](mailto:grevian@gmail.com)