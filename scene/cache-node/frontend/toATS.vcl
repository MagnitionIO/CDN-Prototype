vcl 4.1;

backend edge-backend {
    .host = "edge-backend";
    .port = "8080";
}

sub vcl_recv {
    # Set the backend to the origin server
    set req.backend_hint = edge-backend;
}

sub vcl_backend_fetch {
    set bereq.http.Host = "edge-backend:8080";
}

sub vcl_backend_response {
    if (beresp.http.X-Cache-L1-Store == "True") {
        set beresp.ttl = 60s;
    } else {
        set beresp.ttl = 0s;
    }
}

sub vcl_deliver {
    # Optionally, modify the response headers (e.g., add cache hit/miss headers)
    if (obj.hits > 0) {
        set resp.http.X-Cache-Status = "HIT";
        set resp.http.X-Cache-Node = "VARNISH";
    }
}
