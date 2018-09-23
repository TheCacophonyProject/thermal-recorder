# Introduction

This simple Go package implements a type which represents recurring
daily window between two times of day.

Example:

```go
// Helper to create a time.Time from just hours and minutes.
func mkTime(hour, minute int) time.Time {
    return time.Date(1, 1, 1, hour, minute, 0, 0, time.UTC)
}

// Define a window between 10:10pm and 9:50am.
w := window.New(mkTime(22, 10), mkTime(9, 50))

// At 10pm
w.Active() == false
w.Until() == 10 * time.Minute

// At 10:10pm
w.Active() == true
w.Until() == 0

// At midnight
w.Active() == true
w.Until() == 0

// At 9:49am
w.Active() == true
w.Until() == time.Minute

// At 9:50am
w.Active() == false
w.Until() == 740*time.Minute // Duration until the next window
```

# License

Use of this source code is governed by the Apache License Version 2.0.
