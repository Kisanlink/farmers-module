# Tasks: Farmer ID Foreign Key Implementation

## Status: In Progress

### Phase 1: Schema & Model Changes ✅ COMPLETED

- [x] Analyze current database schema and identify tables needing farmer_id
- [x] Add `farmer_id` field to `Farm` entity with GORM tags
- [x] Add `farmer_id` FK constraint to `CropCycle` entity with GORM relationship
- [x] Add `farmer_id` field to `FarmActivity` entity with GORM relationship
- [x] Update validation methods to require `farmer_id`
- [x] Create migration SQL file (003_add_farmer_id_foreign_keys.sql)
- [x] Verify code compiles successfully
- [x] Document changes in .kiro/specs/farmer-id-foreign-keys/

### Phase 2: Service Layer Updates 🔄 TODO

#### Farm Service
- [ ] Update `CreateFarm` to populate `farmer_id` from farmer lookup
- [ ] Update `UpdateFarm` to validate `farmer_id` consistency
- [ ] Modify queries to use `farmer_id` instead of `aaa_user_id` + `aaa_org_id`
- [ ] Add service methods: `GetFarmsByFarmerID`, `CountFarmsByFarmerID`

#### Crop Cycle Service
- [ ] Update `CreateCropCycle` to populate `farmer_id` from farm
- [ ] Validate `farmer_id` matches farm's `farmer_id` for consistency
- [ ] Update queries to use `farmer_id` for filtering
- [ ] Add preload for `Farmer` relationship where needed

#### Farm Activity Service
- [ ] Update `CreateFarmActivity` to populate `farmer_id` from crop cycle
- [ ] Validate `farmer_id` matches crop cycle's `farmer_id`
- [ ] Update queries to use direct `farmer_id` lookups
- [ ] Add `GetActivitiesByFarmerID` for farmer-level reports

#### Farmer Service
- [ ] Add cascade delete tests when farmer is removed
- [ ] Verify related farms, cycles, and activities are deleted
- [ ] Add transaction safety for batch operations

### Phase 3: Repository Layer Updates 🔄 TODO

#### Farm Repository
- [ ] Add `FindByFarmerID(farmerID string)` method
- [ ] Update `Create` to require `FarmerID`
- [ ] Update `List` to support filtering by `farmer_id`
- [ ] Add index hints for `farmer_id` queries

#### Crop Cycle Repository
- [ ] Update `FindByFarmID` to preload `Farmer` relationship
- [ ] Add `FindByFarmerID` for direct farmer queries
- [ ] Update filtering logic to use `farmer_id`

#### Farm Activity Repository
- [ ] Add `FindByFarmerID` for farmer-level activity queries
- [ ] Update `Create` to require `FarmerID`
- [ ] Add reporting methods using `farmer_id`

### Phase 4: Handler/API Updates 🔄 TODO

#### Farm Handlers
- [ ] Update request validation to ensure `farmer_id` derivation
- [ ] Modify response DTOs if needed
- [ ] Update Swagger annotations

#### Crop Cycle Handlers
- [ ] Ensure `farmer_id` is populated from farm context
- [ ] Update error messages for `farmer_id` validation failures

#### Farm Activity Handlers
- [ ] Ensure `farmer_id` is populated from crop cycle context
- [ ] Add farmer-level activity listing endpoints

### Phase 5: Database Migration 🔄 TODO

- [ ] Review migration SQL for correctness
- [ ] Test migration on sample database
- [ ] Verify data backfill logic
- [ ] Create rollback SQL script
- [ ] Test rollback procedure
- [ ] Run migration in development environment
- [ ] Validate all records have `farmer_id` populated
- [ ] Check foreign key constraints are active

### Phase 6: Testing 🔄 TODO

#### Unit Tests
- [ ] Test `Farm.Validate()` requires `farmer_id`
- [ ] Test `CropCycle.Validate()` requires `farmer_id`
- [ ] Test `FarmActivity.Validate()` requires `farmer_id`

#### Integration Tests
- [ ] Test cascade delete: farmer → farms
- [ ] Test cascade delete: farmer → crop_cycles
- [ ] Test cascade delete: farmer → farm_activities
- [ ] Test FK constraint violation handling
- [ ] Test data integrity with concurrent operations

#### Repository Tests
- [ ] Test `FindByFarmerID` methods
- [ ] Test queries with `farmer_id` filters
- [ ] Test GORM relationship preloading

#### Service Tests
- [ ] Test `farmer_id` population in create operations
- [ ] Test validation of `farmer_id` consistency
- [ ] Test farmer deletion cascade

#### API Tests
- [ ] Test farm creation with valid `farmer_id`
- [ ] Test farm creation with invalid `farmer_id` (should fail)
- [ ] Test crop cycle inherits correct `farmer_id`
- [ ] Test activity inherits correct `farmer_id`

### Phase 7: Performance Testing 🔄 TODO

- [ ] Benchmark queries before and after migration
- [ ] Verify index usage for `farmer_id` columns
- [ ] Test query performance with large datasets
- [ ] Monitor foreign key constraint overhead

### Phase 8: Documentation Updates 🔄 TODO

- [ ] Update API documentation with `farmer_id` fields
- [ ] Update entity relationship diagrams
- [ ] Document migration procedure
- [ ] Update service layer documentation
- [ ] Add inline code comments where needed

### Phase 9: Code Review & QA 🔄 TODO

- [ ] Peer review of entity changes
- [ ] Review migration SQL
- [ ] QA testing of all affected endpoints
- [ ] Security review of cascade delete behavior
- [ ] Performance review of query changes

### Phase 10: Deployment 🔄 TODO

- [ ] Deploy to staging environment
- [ ] Run migration in staging
- [ ] Smoke test all endpoints
- [ ] Monitor logs for errors
- [ ] Verify cascade delete behavior
- [ ] Performance monitoring
- [ ] Deploy to production (with rollback plan ready)
- [ ] Post-deployment verification

## Notes

### Critical Points
1. **Data Migration**: The UPDATE statements in migration must be tested thoroughly
2. **Cascade Delete**: Verify cascade behavior doesn't delete unintended data
3. **Backward Compatibility**: Ensure existing API contracts are maintained
4. **Transaction Safety**: All related updates must be in transactions

### Potential Issues
- Circular import between `farm` and `farmer` packages (resolved by keeping relationship one-directional)
- Existing records without `farmer_id` (handled by migration UPDATE statements)
- Performance impact of foreign key checks (mitigated by indexes)

### Dependencies
- Migration must run before services can use `farmer_id`
- Service updates should be atomic with migration
- Tests should cover both pre- and post-migration states

## Timeline Estimate

- Phase 2-3 (Service & Repository): 2-3 days
- Phase 4 (Handlers/API): 1-2 days
- Phase 5 (Migration): 1 day
- Phase 6 (Testing): 2-3 days
- Phase 7-8 (Performance & Docs): 1 day
- Phase 9-10 (Review & Deploy): 1-2 days

**Total: 8-12 days**

## Assignees

- Backend Engineer: Service/Repository/Handler updates
- DBA: Migration review and execution
- QA Engineer: Testing phases
- DevOps: Deployment and monitoring
