vcl 4.1;

# Define the backend (origin server)
backend origin {
    .host = "origin";  # Use the service name defined in Docker Compose
    .port = "8080";    # Port of the origin server
}

# backend edge-backend {
#     .host = "edge-backend";  # Use the service name defined in Docker Compose
#     .port = "8080";    # Port of the origin server
# }



sub vcl_recv {
    # Set the backend to the origin server
    set req.backend_hint = origin;
}

sub vcl_backend_response {
    # Define conditions to cache the response
    # if (bereq.url ~ "/cacheable") {
    #     set beresp.ttl = 1h;  # Cache for 1 hour (adjust as needed)
    # } else {
    #     set beresp.uncacheable = true;
    #     set beresp.ttl = 1h;  # Cache non-cacheable responses for 2 minutes
    # }
    set beresp.ttl = 1h;
}

sub vcl_deliver {
    # Optionally, modify the response headers (e.g., add cache hit/miss headers)
    if (obj.hits > 0) {
        set resp.http.X-Cache = "HIT";
    } else {
        set resp.http.X-Cache = "MISS";
    }
}
