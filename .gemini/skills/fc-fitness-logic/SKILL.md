---
name: fc-fitness-logic
description: Core domain logic for the Community Fitness Challenge. Use for implementing the Scoring Engine, Anti-cheat heuristics, and Redis-based real-time leaderboard management.
---

# Fitness Domain Logic - Community Fitness Challenge

This skill contains the domain-specific business logic for the application's core functionality: Scoring, Anti-cheat, and Leaderboards.

## 🔢 Scoring Engine

- **Weighted Metrics**: Score = (Calories * 0.4) + (Steps * 0.005) + (ActiveMinutes * 10).
- **Final Score**: The result is multiplied by 100 for storage as an integer.
- **Precision**: Calculations must maintain precision until the final integer conversion to avoid rounding errors.

## 🛡️ Anti-Cheat Heuristics (SQS-based)

The `Anti-Cheat` logic runs asynchronously via the SQS `log_submitted` trigger.
- **Suspicion Score (0-100)**:
  - **Metric Anomalies**: High calories with very few steps or vice versa.
  - **HealthKit Source**: Verification of the source bundle ID and hash.
  - **Impossible Delta**: Significant changes between logs without corresponding time gaps.
- **Thresholds**:
  - `0-30`: Valid log.
  - `31-70`: Marked as suspicious for manual review.
  - `>70`: Automatically rejected; score not applied.

## 🏆 Redis Leaderboards (ZSET)

- **Key Pattern**: `leaderboard:<challenge_id>`.
- **Score Updates**: `ZADD leaderboard:<challenge_id> <total_score> <user_id>`.
- **Leaderboard Types**:
  - **Absolute**: `ZREVRANGE ... WITHSCORES` for top-N ranking.
  - **Relative**: `ZREVRANK` to find the user's position, then `ZREVRANGE` around that rank (e.g., ±2).
- **Concurrency**: Use Redis transactions (`MULTI/EXEC`) or Lua scripts for atomic score increments if needed.

## 📅 Daily Logging Logic

- **Single Submit Lock**: Use a Redis key `log_lock:<user_id>:<date>` with a 24-hour TTL to ensure users can only submit once per day.
- **Timezone Awareness**: The submission date is determined by the user's local timezone, passed in the request.
