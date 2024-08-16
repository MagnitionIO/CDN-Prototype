function read_response()
    local x_cache_store = ts.server_response.header['X-Cache-Store']
    local x_cache_hit_or_miss = ts.server_response.header['X-Cache-Status']

    if ((x_cache_store == "False") or (x_cache_hit_or_miss == "HIT")) then
        ts.http.set_cache_lookup_status (TS_LUA_CACHE_LOOKUP_MISS)
    end
end

function check_cached_response()
    local x_cache_store = ts.cached_response.header['X-Cache-Store']
    local x_cache_hit_or_miss = ts.cached_response.header['X-Cache-Status']

    if ((x_cache_store == "False") or (x_cache_hit_or_miss == "HIT")) then
        ts.http.set_cache_lookup_status (TS_LUA_CACHE_LOOKUP_MISS)
    end
end

function cache_status_check()
    local cache_status = ts.http.get_cache_lookup_status()
    -- ts.debug("Cache status: " .. tostring(cache_status))

    -- local x_cache_store = ts.server_response.header['X-Cache-Store']
    -- local x_cache_hit_or_miss = ts.server_response.header['X-Cache-Status']

    -- if ((x_cache_store == "False") or (x_cache_hit_or_miss == "HIT")) then
    --     txn.server_response.header['Cache-Control'] = 'no-store'
    --     txn.server_response.header['Pragma'] = 'no-cache'
    --     txn.server_response.header['Expires'] = '0'
    -- else
    if cache_status == TS_LUA_CACHE_LOOKUP_HIT_FRESH then
        ts.client_response.header['X-Cache-Status'] = 'HIT'
        ts.client_response.header['X-Cache-Node'] = 'ATS'
    end
end

-- ts.hook(TS_LUA_HOOK_READ_RESPONSE_HDR, read_response)
-- ts.hook(TS_LUA_HOOK_CACHE_LOOKUP_COMPLETE, check_cached_response)
ts.hook(TS_LUA_HOOK_SEND_RESPONSE_HDR, cache_status_check)