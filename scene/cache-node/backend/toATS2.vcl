vcl 4.1;

backend origin {
    .host = "origin";
    .port = "8080";
}

sub vcl_recv {
    set req.backend_hint = origin;
}

sub vcl_backend_fetch {
    set bereq.http.Host = "origin:8080";
}

sub vcl_backend_response {
    if (beresp.http.X-Cache-L2-Store == "True") {
        set beresp.ttl = 365d;
    } else {
        set beresp.ttl = 0s;
    }
}

sub vcl_deliver {
    # Optionally, modify the response headers (e.g., add cache hit/miss headers)
    if (obj.hits > 0) {
        set resp.http.X-Cache-Status = "HIT";
        set resp.http.X-Cache-Node = "L2-2";
    }
}
