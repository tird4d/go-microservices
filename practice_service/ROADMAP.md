# Go Practice Roadmap 🚀

A step-by-step guide to learn Go from basics to advanced concepts.

---

## Phase 1: Fundamentals ✅

### Lesson 1: Variables & Data Types
- [ ] Variable declaration (`var`, `:=`)
- [ ] Basic types: `string`, `int`, `float64`, `bool`
- [ ] Zero values
- [ ] Constants (`const`)
- [ ] Type conversion

### Lesson 2: Control Flow
- [ ] `if`, `else`, `else if`
- [ ] `switch` statement
- [ ] `for` loops (Go's only loop!)
- [ ] `break`, `continue`

### Lesson 3: Functions
- [ ] Function declaration
- [ ] Multiple return values
- [ ] Named return values
- [ ] Variadic functions (`...`)
- [ ] Anonymous functions

### Lesson 4: Collections
- [ ] Arrays
- [ ] Slices (very important!)
- [ ] Maps
- [ ] `range` keyword

---

## Phase 2: Intermediate 🔧

### Lesson 5: Structs
- [ ] Struct declaration
- [ ] Struct fields
- [ ] Nested structs
- [ ] Anonymous structs

### Lesson 6: Methods & Pointers
- [ ] Pointers (`*`, `&`)
- [ ] Methods on structs
- [ ] Value receivers vs pointer receivers

### Lesson 7: Interfaces
- [ ] Interface declaration
- [ ] Implementing interfaces
- [ ] Empty interface (`interface{}` / `any`)
- [ ] Type assertions
- [ ] Type switch

### Lesson 8: Error Handling
- [ ] The `error` type
- [ ] Creating custom errors
- [ ] Error wrapping
- [ ] `panic` and `recover`

---

## Phase 3: Advanced 🚀

### Lesson 9: Concurrency - Goroutines
- [ ] Goroutines (`go` keyword)
- [ ] `sync.WaitGroup`
- [ ] `sync.Mutex`

### Lesson 10: Concurrency - Channels
- [ ] Channel basics
- [ ] Buffered vs unbuffered channels
- [ ] `select` statement
- [ ] Channel patterns

### Lesson 11: Context
- [ ] `context.Context`
- [ ] Cancellation
- [ ] Timeouts
- [ ] Passing values

### Lesson 12: Testing
- [ ] Unit tests
- [ ] Table-driven tests
- [ ] Benchmarks
- [ ] Mocking

---

## Phase 4: Real-World Patterns 🌐

### Lesson 13: HTTP Server
- [ ] `net/http` package
- [ ] Handlers
- [ ] Middleware pattern
- [ ] JSON encoding/decoding

### Lesson 14: Working with Databases
- [ ] SQL basics
- [ ] MongoDB driver
- [ ] Repository pattern

### Lesson 15: gRPC
- [ ] Protocol Buffers
- [ ] gRPC server
- [ ] gRPC client

### Lesson 16: Building a Microservice
- [ ] Project structure
- [ ] Configuration management
- [ ] Logging
- [ ] Graceful shutdown
- [ ] Docker containerization

---

## Progress Tracker

| Phase | Status | Completed |
|-------|--------|-----------|
| Phase 1 | 🟡 In Progress | 0/4 |
| Phase 2 | ⚪ Not Started | 0/4 |
| Phase 3 | ⚪ Not Started | 0/4 |
| Phase 4 | ⚪ Not Started | 0/4 |

---

## Files Structure

```
practice_service/
├── ROADMAP.md           # This file
├── main.go              # Current lesson
├── go.mod               # Module file
└── lessons/             # (Future) Separate lesson files
    ├── 01_variables/
    ├── 02_control_flow/
    ├── 03_functions/
    └── ...
```

---

**Let's Go! 🎯**
