# Simple BFF Example

This is a minimal example that demonstrates the core concepts of building a BFF (Backend for Frontend) using the API Orchestration Framework.

## What This Example Shows

1. **Basic Flow Creation** - How to create and execute flows
2. **Context Propagation** - How data flows between steps
3. **Caching Pattern** - Check cache first, fetch if needed
4. **Parallel Execution** - Fetch multiple APIs simultaneously
5. **Data Transformation** - Convert data for mobile consumption
6. **Conditional Logic** - Different paths based on conditions

## How to Run

1. **Install dependencies:**
   ```bash
   go mod init simple-bff-example
   go get github.com/gin-gonic/gin
   go get go.uber.org/zap
   go get github.com/venkatvghub/api-orchestration-framework
   ```

2. **Run the server:**
   ```bash
   go run main.go
   ```

3. **Test the endpoints:**
   ```bash
   # Health check
   curl http://localhost:8080/health
   
   # Get user (with caching)
   curl http://localhost:8080/users/1
   
   # Get dashboard (parallel fetching)
   curl http://localhost:8080/dashboard/1
   ```

## Understanding the Code

### Example 1: User with Caching

```go
userFlow.
    // Step 1: Check cache
    Step("check_cache", checkCacheStep).
    
    // Step 2: Conditional fetch
    Choice("cache_status").
        When(func(ctx interfaces.ExecutionContext) bool {
            return !ctx.Has("cached_user")
        }).
            Step("fetch_user", fetchFromAPI).
            Step("cache_user", storeInCache).
        Otherwise().
            Step("use_cached", useCachedData).
    EndChoice().
    
    // Step 3: Transform for mobile
    Transform("mobile_transform", optimizeForMobile)
```

**What happens:**
1. Check if user is in cache
2. If not in cache → fetch from API and store in cache
3. If in cache → use cached data
4. Transform data to mobile-friendly format

### Example 2: Dashboard with Parallel Fetching

```go
dashboardFlow.
    // Fetch multiple things at once (FAST!)
    Parallel("fetch_dashboard_data").
        Step("user", fetchUser).
        Step("posts", fetchPosts).
        Step("todos", fetchTodos).
    EndParallel().
    
    // Combine all the data
    Transform("combine_data", combineAllData)
```

**What happens:**
1. Fetch user, posts, and todos **simultaneously** (not one by one)
2. Combine all the data into a dashboard response
3. Return optimized data to mobile

## Key Concepts Demonstrated

### 1. Context Propagation
```go
// Set data in context
ctx.Set("user_id", userID)

// Get data from context
userID, _ := ctx.GetString("user_id")

// Check if data exists
if ctx.Has("cached_user") {
    // Use cached data
}
```

### 2. Flow Steps
- **Sequential**: Steps run one after another
- **Parallel**: Steps run at the same time
- **Conditional**: Different paths based on conditions
- **Transform**: Change data format

### 3. Error Handling
```go
_, err := userFlow.Execute(flowCtx)
if err != nil {
    c.JSON(500, APIResponse{
        Success: false,
        Error:   err.Error(),
    })
    return
}
```

### 4. Mobile Optimization
```go
// Instead of sending everything:
{
  "id": 1,
  "name": "John Doe",
  "username": "johndoe",
  "email": "john@example.com",
  "address": { /* lots of data */ },
  "phone": "123-456-7890",
  "website": "johndoe.com",
  "company": { /* more data */ }
}

// Send only what mobile needs:
{
  "id": 1,
  "name": "John Doe"
}
```

## Why This Approach is Better

### Without BFF (Traditional Approach)
```
Mobile App:
1. Call /users/1        (200ms)
2. Call /users/1/posts  (300ms)  
3. Call /users/1/todos  (250ms)
Total: 750ms + network overhead
```

### With BFF (This Example)
```
Mobile App:
1. Call /dashboard/1    (300ms - all APIs called in parallel)
Total: 300ms + optimized data
```

**Result: 2.5x faster + less data transferred!**

## Next Steps

After understanding this example:

1. **Add more endpoints** - Try creating a posts endpoint
2. **Add real caching** - Use Redis instead of in-memory cache
3. **Add authentication** - Validate user tokens
4. **Add error handling** - Handle API failures gracefully
5. **Add metrics** - Track performance and errors

## Common Patterns You'll Use

1. **Cache-First**: Always check cache before API
2. **Parallel Fetch**: Get multiple things at once
3. **Fallback**: If primary API fails, use backup
4. **Transform**: Convert backend data to mobile format
5. **Validate**: Check inputs before processing

This example gives you the foundation to build more complex BFF layers! 