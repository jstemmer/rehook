# rehook

![](https://github.com/gophergala/rehook/blob/master/public/images/rehook-logo.png)

__Rehook__ - a webhook dispatcher, filtering incoming requests from external
services and acting on them.

Initial version created during the [Gopher Gala](http://gophergala.com).

[![Build Status](https://travis-ci.org/gophergala/rehook.svg?branch=master)](https://travis-ci.org/gophergala/rehook)

## About

External web services often provide webhooks as an alternative to polling their
public API's for up-to-date information. For example, commits being pushed to
Github, bounced emails when sending automated messages through email services,
or chat messages matching a certain filter.

However, you don't want to deal with all the different formats, validations and
building filters for each of those services on every system that is interested
in this information. Some systems might not even be directly reachable from the
internet.

__Rehook__ provides a central, public web endpoint that receives all those
webhook requests. It can validate them to prevent abuse, filter them for
certain conditions or rate-limit incoming requests. You can configure actions
that should take place afterwards, such as forwarding the request to multiple
hosts, storing them for later analysis or sending an email based on a template.

It provides an easy to use web interface to configure your webhooks and to keep
an eye on how much traffic is being handled.

![](https://github.com/gophergala/rehook/blob/master/screenshots/rehook_stats.png)

[More screenshots](https://github.com/gophergala/rehook/tree/master/screenshots)

## Building from source

Assuming you have installed a recent version of
[Go](https://golang.org/doc/install), you can simply run 

```
go get github.com/gophergala/rehook
```

This will download Rehook to `$GOPATH/src/github.com/gophergala/rehook`. In
this directory run `go build` to create the `rehook` binary.

## Running

Simply start the `rehook` binary in the directory where you downloaded Rehook.
When running the binary from elsewhere, make sure to copy the `public` and
`views` directories, as they contain the files required for the admin web
interface.

By default, Rehook will listen to
[http://localhost:9000](http://localhost:9000) for incoming webhook requests.
The admin interface where you can configure Rehook is available on
[http://localhost:9001/](http://localhost:9000).

You should make sure to prevent unauthorized access to the administration port.
In the future we will replace this with proper user management and
authentication.

You can configure these ports and the location of the database file.

```
$ ./rehook -help
Usage of ./rehook:
  -admin=":9001": Private HTTP listen address for admin interface
  -db="data.db": Database file to use
  -http=":9000": Public HTTP listen address for incoming webhooks
```

## Configuring your first webhook

Open the admin interface in your browser,
[http://localhost:9001](http://localhost:9001) by default.

Every hook you create must have a unique identifier, Rehook will listen on the
public port for incoming requests on the `/h/<identifier>` path.

Click the `Create new hook` button to create a new webhook. Enter a suitable
hook identifier, for example `github`. Rehook will now accept HTTP requests on
`http://localhost:9000/h/github`.

You can now add components to process incoming requests. The arrow indicates
the order in which the request flows through the components. If a component
cannot handle a request, it will give a reason an further processing is
stopped.

Since this webhook will accept Github webhook requests, let's first add a
`Github validator` component. This component makes sure that the incoming
request is actually from Github by calculating the signature of the request
data and comparing it to the signature in the request header. The secret you
enter here must be the same as the secret you will enter on the Github webhook
page.

Let's log the request and write it to a file. Add the `Log` and `Write to file`
components. The last one will write a file for each request in the `log/`
directory. Your page should now [look something like this](https://github.com/gophergala/rehook/blob/master/screenshots/rehook_chain.png).

Go to the Github settings page of your repository, click the `Webhooks &
Services` button and add a new Webhook. Enter the URL where your Rehook
instance can be reached, be sure to use `/h/github` as the path. Click `Add
webhook` and if everything was setup correctly, your first webhook request was
just handled by Rehook.

## Components

The following components are currently available:

### Send email (using Mailgun)

Sends an email with a custom body template to the specified address using the
Mailgun API.

### Forward request

Forwards the request, including its headers, to the specified URL.

### Github validator

Calculates the SHA1 HMAC of the body and compares it to the `X-Hub-Signature`
header. In addition, it makes sure the `X-Github-Delivery` header is unique to
prevent replay attacks.

### Log

Logs a message to `stderr`.

### Mailgun validator

Calculates the SHA256 HMAC of the `timestamp` and random `token` in the request
body and compares it to the `signature`. It also verifies the `token` is unique
to prevent replay attacks.

### Rate limiter

The rate limiter accepts a certain number of requests in a configurable
interval. Incoming requests exceeding this limit will be dropped.

### Write to file

Writes the contents of the request to a file in the `log/` directory. This
makes it easy to view the request details later.

## License

MIT, see the LICENSE file.
