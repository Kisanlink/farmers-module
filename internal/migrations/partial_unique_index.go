package migrations

import "gorm.io/gorm"

// MigratePartialUniqueIndex replaces the standard unique index on farmers.aaa_user_id
// with a partial unique index that only covers non-deleted records.
// This allows re-adding a farmer with the same aaa_user_id after soft deletion.
func MigratePartialUniqueIndex(db *gorm.DB) error {
	// Drop the old full unique index if it exists
	if db.Migrator().HasIndex("farmers", "idx_farmer_aaa_user_id") {
		if err := db.Migrator().DropIndex("farmers", "idx_farmer_aaa_user_id"); err != nil {
			return err
		}
	}

	// Create a partial unique index that only applies to non-deleted records
	return db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS idx_farmer_aaa_user_id
                    ON farmers(aaa_user_id) WHERE deleted_at IS NULL`).Error
}
