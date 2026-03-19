---
name: fc-ios-swift-fitness
description: Senior SwiftUI developer for the Community Fitness Challenge app. Use for implementing iOS 17+ features, HealthKit integration, SwiftData caching, and SPM modular architecture.
---

# iOS Senior Development - Community Fitness Challenge

This skill defines the standards for the "Comm Fit Challenge" iOS app. Adhere to these senior-level Swift/SwiftUI practices.

## 🏗️ Architecture & Modularity (SPM)

The app is modularized via Swift Package Manager (SPM).
- **`FCDesignSystem`**: All UI components (FCButton, FCCard, FCProgressBar). Monochromatic style (Black/White/Gray).
- **`FCNetworking`**: Generic `APIClient` using `async/await`. Interceptors for JWT auth and retry logic.
- **`FCHealthKit`**: `HealthKitManager` (Actor) for all health-related queries.
- **`FCPersistence`**: SwiftData models for offline caching (CachedChallenge, CachedLog).

## 🏃 HealthKit & Metrics

- **Permissions**: Always check `HKHealthStore.isHealthDataAvailable()` and handle permission flows gracefully.
- **Metric Mapping**:
  - `HKQuantityType.quantityType(forIdentifier: .stepCount)`
  - `HKQuantityType.quantityType(forIdentifier: .activeEnergyBurned)`
  - `HKQuantityType.quantityType(forIdentifier: .appleExerciseTime)`
- **Daily Sync**: Use the `DailyLogViewModel` to sync and hash HealthKit data before submission.

## 📱 SwiftUI Standards

- **State Management**: Use `@State`, `@Binding`, `@ObservedObject` (or `@Observable` for iOS 17+).
- **Themes**: Strict adherence to the monocolor design system.
- **Navigation**: Use `NavigationStack` with typed routes.
- **Accessibility**: VoiceOver labels and Dynamic Type support are mandatory for all UI components.

## 🔐 Auth & Tokens

- **Apple Sign-In**: Use `ASAuthorizationAppleIDButton`. Store `apple_user_id` and JWT in the secure Keychain.
- **Token Handling**: Automatic token refresh via networking interceptors. Log out on 401.

## ⚡ Offline & Real-time

- **SwiftData**: Cache API results for offline viewing. Implement "Pending Upload" for logs submitted while offline.
- **WebSockets**: Live leaderboard updates with fluid animations for rank changes.
