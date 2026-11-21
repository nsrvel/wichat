package authRedis

import (
	"context"

	"github.com/redis/go-redis/v9"
	"github.com/wichat/wichat/backend/ms-user/pkg/constants"
)

// verifyRegisterChallengeLua checks register OTP but keeps the pending payload until the caller deletes the key.
// On correct code it sets otp_verified=true so the same code can retry user creation after a transient failure.
const verifyRegisterChallengeLua = `
local raw = redis.call('GET', KEYS[1])
if not raw then return {0} end
local ok, rec = pcall(cjson.decode, raw)
if not ok then return {0} end
if rec['otp_verified'] == true then
  if rec['code_hmac'] == ARGV[1] then
    return {1, raw}
  end
  return {0}
end
if rec['code_hmac'] == ARGV[1] then
  rec['otp_verified'] = true
  local enc = cjson.encode(rec)
  local ttl = redis.call('TTL', KEYS[1])
  if ttl > 0 then
    redis.call('SET', KEYS[1], enc, 'EX', ttl)
  else
    redis.call('SET', KEYS[1], enc, 'EX', 600)
  end
  return {1, enc}
end
rec['attempts'] = (rec['attempts'] or 0) + 1
if rec['attempts'] >= tonumber(ARGV[2]) then
  redis.call('DEL', KEYS[1])
  return {0}
end
local enc = cjson.encode(rec)
local ttl = redis.call('TTL', KEYS[1])
if ttl > 0 then
  redis.call('SET', KEYS[1], enc, 'EX', ttl)
else
  redis.call('SET', KEYS[1], enc, 'EX', 600)
end
return {0}
`

var verifyRegisterChallengeScript = redis.NewScript(verifyRegisterChallengeLua)

// VerifyRegisterChallengeJSON atomically verifies a register OTP without deleting the challenge key.
func VerifyRegisterChallengeJSON(ctx context.Context, rdb *redis.Client, key, submittedHMAC string, maxAttempts int) (okPayload []byte, err error) {
	res, err := verifyRegisterChallengeScript.Run(ctx, rdb, []string{key}, submittedHMAC, maxAttempts).Result()
	if err != nil {
		return nil, err
	}
	arr, ok := res.([]any)
	if !ok || len(arr) == 0 {
		return nil, constants.ErrNotFound
	}
	switch n, _ := arr[0].(int64); n {
	case 1:
		if len(arr) < 2 {
			return nil, constants.ErrNotFound
		}
		switch payload := arr[1].(type) {
		case string:
			return []byte(payload), nil
		case []byte:
			return payload, nil
		default:
			return nil, constants.ErrNotFound
		}
	default:
		return nil, constants.ErrNotFound
	}
}
