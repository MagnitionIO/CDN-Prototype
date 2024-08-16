vcl 4.1;

backend edge-backend {
    .host = "edge-backend";
    .port = "8080";
}

backend edge-backend-2 {
    .host = "edge-backend-2";
    .port = "8080";
}

sub vcl_recv {
    if (req.http.X-Cache-L2-Server == "0") {
        set req.backend_hint = edge-backend;
    } else {
        set req.backend_hint = edge-backend-2;
    }
}

sub vcl_backend_fetch {
#    set bereq.http.Host = "edge-backend:8080";
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
