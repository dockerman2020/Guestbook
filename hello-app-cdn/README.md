# Hello Application with CDN example

> **Note:** This application is a copy of [hello-app](../hello-app) sample.
> See that directory for more details on this sample.

This sample web application is designed to be compatible with the Cloud CDN.
It responds the requests with the `Cache-Control` HTTP header to ensure the responses
are cached.

The container image for this directory is publicly available at
`gcr.io/google-samples/hello-app-cdn:1.0`.
