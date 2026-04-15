package domain

// Redis key formats
const (
	RedisKeyChallengeCount = "challenge_count:%s"
	RedisKeyRateLimit      = "rl:%s"
	RedisKeyDailyLogLock   = "daily_log_lock:%s:%s:%s"
	RedisKeyLeaderboard    = "leaderboard:%s"
)
