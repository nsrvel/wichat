package authRedis

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/redis/go-redis/v9"
	"github.com/wichat/wichat/backend/ms-user/pkg/constants"
)

// verifyChallengeLua atomically checks code_hmac, bumps attempts, or deletes the key on success.
// Returns Redis array: {1, payload} on success (key removed), {0} on wrong/missing/exhausted.
const verifyChallengeLua = `
local raw = redis.call('GET', KEYS[1])
if not raw then return {0} end
local ok, rec = pcall(cjson.decode, raw)
if not ok then return {0} end
if rec['code_hmac'] == ARGV[1] then
  redis.call('DEL', KEYS[1])
  return {1, raw}
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

var verifyChallengeScript = redis.NewScript(verifyChallengeLua)

// VerifyChallengeJSON runs the atomic OTP verify script. okPayload is the previous JSON blob when verified.
func VerifyChallengeJSON(ctx context.Context, rdb *redis.Client, key, submittedHMAC string, maxAttempts int) (okPayload []byte, err error) {
	res, err := verifyChallengeScript.Run(ctx, rdb, []string{key}, submittedHMAC, maxAttempts).Result()
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

// DecodeJSON unmarshals challenge payload bytes into dst.
func DecodeJSON(payload []byte, dst any) error {
	if len(payload) == 0 {
		return constants.ErrNotFound
	}
	if err := json.Unmarshal(payload, dst); err != nil {
		return err
	}
	return nil
}

// IsVerifyNotFound reports whether err means invalid, exhausted, or missing challenge.
func IsVerifyNotFound(err error) bool {
	return errors.Is(err, constants.ErrNotFound)
}
