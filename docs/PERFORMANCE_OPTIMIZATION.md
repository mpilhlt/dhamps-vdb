# Performance Optimization Notes

## Query Optimization Opportunities

### GetAllAccessibleInstances Query

**Current Implementation:**
```sql
SELECT instances.*, ...
FROM instances
LEFT JOIN instances_shared_with
  ON instances."instance_id" = instances_shared_with."instance_id"
WHERE instances."owner" = $1
   OR instances_shared_with."user_handle" = $1
ORDER BY instances."owner" ASC, instances."instance_handle" ASC 
LIMIT $2 OFFSET $3;
```

**Issue:**
The LEFT JOIN combined with OR conditions in WHERE clause may result in inefficient query execution. The query planner might struggle to use indexes effectively.

**Recommended Optimization:**
Use UNION ALL to separate owned instances from shared instances:

```sql
-- Get owned instances
SELECT instances.*, 'owner' as "role", true as "is_owner"
FROM instances
WHERE instances."owner" = $1

UNION ALL

-- Get shared instances
SELECT instances.*, 
       instances_shared_with."role",
       false as "is_owner"
FROM instances
INNER JOIN instances_shared_with
  ON instances."instance_id" = instances_shared_with."instance_id"
WHERE instances_shared_with."user_handle" = $1
  AND instances."owner" != $1  -- Avoid duplicates

ORDER BY "owner" ASC, "instance_handle" ASC
LIMIT $2 OFFSET $3;
```

**Benefits:**
1. Query planner can use separate index scans for each UNION branch
2. Owned instances can use index on (owner)
3. Shared instances can use index on (user_handle)
4. Clearer query execution plan
5. Better performance with large datasets

**Tradeoff:**
- Slightly more complex SQL
- Need to deduplicate if user somehow has instance both owned and shared (unlikely scenario)

**Recommendation:**
- Current implementation is correct and works well for small-medium datasets
- Consider optimization if performance becomes an issue with large numbers of instances
- Profile with EXPLAIN ANALYZE before and after optimization

## Other Optimization Opportunities

### Index Suggestions

Current indexes (from migration 004):
- `definitions(definition_handle)`
- `definitions(owner, definition_handle)` (composite)
- `instances(instance_handle)`
- `instances_shared_with(instance_id, user_handle)` (implicit from PK)

**Additional indexes to consider:**
1. `instances(owner)` - for owned instance lookups
2. `instances_shared_with(user_handle)` - for shared instance lookups
3. `instances(owner, instance_handle)` - composite for unique constraint

### Caching Opportunities

1. **System Definitions**: Cache _system definitions since they rarely change
2. **User Instances**: Cache user's instance list with short TTL
3. **API Standards**: Cache list of API standards (nearly static)

### Query Analysis Tools

```bash
# Analyze query performance
EXPLAIN ANALYZE SELECT ...;

# Check table statistics
ANALYZE instances;
ANALYZE instances_shared_with;

# View current indexes
\di llm_service_*
```

## Performance Testing

### Recommended Tests

1. **Load Test**: 1000 users, 10 instances each
2. **Sharing Test**: 100 users sharing instances with 50 others each
3. **Query Test**: Measure GetAllAccessibleInstances with varying instance counts

### Metrics to Track

- Query execution time (p50, p95, p99)
- Database connection pool usage
- Index hit rates
- Cache hit rates (if implemented)

### Performance Targets

Based on typical usage:
- Single instance lookup: < 10ms
- List all accessible instances: < 50ms (for < 100 instances)
- Create/update instance: < 100ms (including encryption)

## Implementation Priority

1. **High**: Profile current performance with realistic data
2. **Medium**: Implement UNION ALL optimization if query time > 100ms
3. **Low**: Add caching layer for frequently accessed data
4. **Low**: Add indexes based on actual query patterns

## Notes

- Current implementation prioritizes correctness over optimization
- All tests pass with current query structure
- Performance optimization should be data-driven (measure first)
- Don't optimize prematurely - wait for actual performance issues
