# go-get-redirect ![Test run](https://github.com/gentlecat/publisher/workflows/Test%20run/badge.svg)

This tool allows you to vend your [Go](https://go.dev/) packages via a custom domain name. It generates a static website for redirection purposes by scanning a given GitHub account for public Go repos. Current implementation is based around GitHub account and generating a static website using [Actions](https://github.com/features/actions).

This is useful when you don't want to run a server that does the redirection. If you do have a server to run this on, then you could do something like this with [nginx](https://nginx.org):

```
location ~ ^/([^/]+).*$ {
	if ($args = "go-get=1") {
		add_header Content-Type text/html;
		return 200 '<meta name="go-import" content="$host/$1 git https://github.com/USERNAME/$1.git">';
	}
	return 302 https://github.com/USERNAME/$1;
}
```

Additional information about how this works can be found [here](https://pkg.go.dev/cmd/go#hdr-Remote_import_paths).
