package limiter

// Request:
// KEYS[1] -> The unique key (e.g., "rate_limit:192.168.1.1")
// ARGV[1] -> Capacity (Max burst allowed, e.g., 10)
// ARGV[2] -> Refill Rate (Tokens added per second, e.g., 1.0)
// ARGV[3] -> Current Timestamp (Unix seconds, passed from Go)
// ARGV[4] -> Cost (How many tokens this request costs, usually 1)

const requestScript = `
local key = KEYS[1]
local capacity = tonumber(ARGV[1])
local rate = tonumber(ARGV[2])
local now = tonumber(ARGV[3])
local cost = tonumber(ARGV[4])

-- Retrieve current state: [tokens_left, last_refill_timestamp]
local info = redis.call("HMGET", key, "tokens", "last_refill")
local tokens = tonumber(info[1])
local last_refill = tonumber(info[2])

-- Initialize if key doesn't exist
if not tokens then
    tokens = capacity
    last_refill = now
end

-- CALCULATION: How many tokens regenerated since last time?
-- Formula: (Now - Last_Refill) * Rate
local delta = math.max(0, now - last_refill)
local to_add = delta * rate

-- Refill the bucket (but don't exceed capacity)
tokens = math.min(capacity, tokens + to_add)

-- CHECK: Do we have enough tokens?
if tokens < cost then
    return 0 -- Rejected
end

-- DEDUCT: Pay the cost
tokens = tokens - cost

-- SAVE: Update state atomically
redis.call("HMSET", key, "tokens", tokens, "last_refill", now)

-- EXPIRE: Clean up key if user stops (Time to refill full bucket * 2 safety factor)
local ttl = math.ceil(capacity / rate) * 2
redis.call("EXPIRE", key, ttl)

return 1 -- Allowed`
