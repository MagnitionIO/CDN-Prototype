vcl 4.1;

backend edge-backend-1 {
    .host = "edge-backend-1";
    .port = "80";
}

backend edge-backend-2 {
    .host = "edge-backend-2";
    .port = "80";
}

backend edge-backend-3 {
    .host = "edge-backend-3";
    .port = "80";
}

backend edge-backend-4 {
    .host = "edge-backend-4";
    .port = "80";
}

acl purge {
    "172.18.0.0/24";
}

sub vcl_recv {
    if (req.method == "PURGE") {
        if (client.ip ~ purge) {
            return (purge);
        }
    }

    if (req.http.X-Cache-L2-Server == "0") {
        set req.backend_hint = edge-backend-1;
    } else if (req.http.X-Cache-L2-Server == "1") {
        set req.backend_hint = edge-backend-2;
    } else if (req.http.X-Cache-L2-Server == "2") {
        set req.backend_hint = edge-backend-3;
    } else if (req.http.X-Cache-L2-Server == "3") {
        set req.backend_hint = edge-backend-4;
    } else {
        set req.backend_hint = edge-backend-1;
    }
}

sub vcl_backend_fetch {
    if (bereq.http.X-Cache-L2-Server == "0") {
        set bereq.http.Host = "edge-backend-1:80";
    } else if (bereq.http.X-Cache-L2-Server == "1") {
        set bereq.http.Host = "edge-backend-2:80";
    } else if (bereq.http.X-Cache-L2-Server == "2") {
        set bereq.http.Host = "edge-backend-3:80";
    } else if (bereq.http.X-Cache-L2-Server == "3") {
        set bereq.http.Host = "edge-backend-4:80";
    } else {
        set bereq.http.Host = "edge-backend-1:80";
    }
#    set bereq.http.Host = "edge-backend:8080";
}

sub vcl_backend_response {
    if (beresp.http.X-Cache-L1-Store == "False" || beresp.http.X-Cache-Status == "HIT") {
        set beresp.ttl = 0s;
    } else {
        set beresp.ttl = 365d;
    }
}

sub vcl_deliver {
    # Optionally, modify the response headers (e.g., add cache hit/miss headers)
    if (obj.hits > 0) {
        set resp.http.X-Cache-Status = "HIT";
        set resp.http.X-Cache-Node = "L1-3";
        set resp.http.X-Cache-Hit-Count = obj.hits;
    }
}