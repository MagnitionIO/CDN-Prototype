function cache_status_check()
    local cache_status = ts.http.get_cache_lookup_status()
    ts.debug("Cache status: " .. tostring(cache_status))

    if cache_status == TS_LUA_CACHE_LOOKUP_HIT_FRESH then
        ts.client_response.header['X-Cache-Status'] = 'HIT'
        ts.client_response.header['X-Cache-Node'] = 'ATS'
    end
end

ts.hook(TS_LUA_HOOK_SEND_RESPONSE_HDR, cache_status_check)