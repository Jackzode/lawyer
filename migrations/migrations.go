package migrations

import (
	"context"
	"fmt"
	"github.com/lawyer/commons/entity"
	"github.com/lawyer/commons/handler"

	"xorm.io/xorm"
)

const minDBVersion = 0

// Migration describes on migration from lower version to high version
type Migration interface {
	Version() string
	Description() string
	Migrate(ctx context.Context, x *xorm.Engine) error
	ShouldCleanCache() bool
}

type migration struct {
	version          string
	description      string
	migrate          func(ctx context.Context, x *xorm.Engine) error
	shouldCleanCache bool
}

// Version returns the migration's version
func (m *migration) Version() string {
	return m.version
}

// Description returns the migration's description
func (m *migration) Description() string {
	return m.description
}

// Migrate executes the migration
func (m *migration) Migrate(ctx context.Context, x *xorm.Engine) error {
	return m.migrate(ctx, x)
}

// ShouldCleanCache should clean the cache
func (m *migration) ShouldCleanCache() bool {
	return m.shouldCleanCache
}

// NewMigration creates a new migration
func NewMigration(version, desc string, fn func(ctx context.Context, x *xorm.Engine) error, shouldCleanCache bool) Migration {
	return &migration{version: version, description: desc, migrate: fn, shouldCleanCache: shouldCleanCache}
}

// Use noopMigration when there is a migration that has been no-oped
var noopMigration = func(_ context.Context, _ *xorm.Engine) error { return nil }

var migrations = []Migration{
	// 0->1
	NewMigration("v0.0.1", "this is first version, no operation", noopMigration, false),
	NewMigration("v0.3.0", "add user language", addUserLanguage, false),
	NewMigration("v0.4.1", "add recommend and reserved tag fields", addTagRecommendedAndReserved, false),
	NewMigration("v0.5.0", "add activity timeline", addActivityTimeline, false),
	NewMigration("v0.6.0", "add user role", addRoleFeatures, false),
	NewMigration("v1.0.0", "add theme and private mode", addThemeAndPrivateMode, true),
	NewMigration("v1.0.2", "add new answer notification", addNewAnswerNotification, true),
	NewMigration("v1.0.5", "add plugin", addPlugin, false),
	NewMigration("v1.0.7", "add user pin hide features", addRolePinAndHideFeatures, true),
	NewMigration("v1.0.8", "update accept answer rank", updateAcceptAnswerRank, true),
	NewMigration("v1.0.9", "add login limitations", addLoginLimitations, true),
	NewMigration("v1.1.0-beta.1", "update user pin hide features", updateRolePinAndHideFeatures, true),
	NewMigration("v1.1.0-beta.2", "update question post time", updateQuestionPostTime, true),
	NewMigration("v1.1.0", "add gravatar base url", updateCount, true),
	NewMigration("v1.1.1", "update the length of revision content", updateTheLengthOfRevisionContent, false),
	NewMigration("v1.1.2", "add notification config", addNoticeConfig, true),
	NewMigration("v1.1.3", "set default user notification config", setDefaultUserNotificationConfig, false),
	NewMigration("v1.2.0", "add recover answer permission", addRecoverPermission, true),
	NewMigration("v1.2.1", "add password login control", addPasswordLoginControl, true),
}

func GetMigrations() []Migration {
	return migrations
}

// GetCurrentDBVersion returns the current db version
func GetCurrentDBVersion(engine *xorm.Engine) (int64, error) {
	if err := engine.Sync(new(entity.Version)); err != nil {
		return -1, fmt.Errorf("sync version failed: %v", err)
	}

	currentVersion := &entity.Version{ID: 1}
	has, err := engine.Get(currentVersion)
	if err != nil {
		return -1, fmt.Errorf("get first version failed: %v", err)
	}
	if !has {
		_, err := engine.InsertOne(&entity.Version{ID: 1, VersionNumber: 0})
		if err != nil {
			return -1, fmt.Errorf("insert first version failed: %v", err)
		}
		return 0, nil
	}
	return currentVersion.VersionNumber, nil
}

// ExpectedVersion returns the expected db version
func ExpectedVersion() int64 {
	return int64(minDBVersion + len(migrations))
}

// Migrate database to current version
func Migrate(debug bool, upgradeToSpecificVersion string) error {

	currentDBVersion, err := GetCurrentDBVersion(handler.Engine)
	if err != nil {
		return err
	}
	expectedVersion := ExpectedVersion()
	if len(upgradeToSpecificVersion) > 0 {
		fmt.Printf("[migrate] user set upgrade to version: %s\n", upgradeToSpecificVersion)
		for i, m := range migrations {
			if m.Version() == upgradeToSpecificVersion {
				currentDBVersion = int64(i)
				break
			}
		}
	}

	for currentDBVersion < expectedVersion {
		fmt.Printf("[migrate] current db version is %d, try to migrate version %d, latest version is %d\n",
			currentDBVersion, currentDBVersion+1, expectedVersion)
		migrationFunc := migrations[currentDBVersion]
		fmt.Printf("[migrate] try to migrate Answer version %s, description: %s\n", migrationFunc.Version(), migrationFunc.Description())
		if err := migrationFunc.Migrate(context.Background(), handler.Engine); err != nil {
			fmt.Printf("[migrate] migrate to db version %d failed: %s\n", currentDBVersion+1, err.Error())
			return err
		}
		if migrationFunc.ShouldCleanCache() {
			//if err := handler.RedisClient.Flush(context.Background()); err != nil {
			//	fmt.Printf("[migrate] flush cache failed: %s\n", err.Error())
			//}
		}
		fmt.Printf("[migrate] migrate to db version %d success\n", currentDBVersion+1)
		if _, err := handler.Engine.Update(&entity.Version{ID: 1, VersionNumber: currentDBVersion + 1}); err != nil {
			fmt.Printf("[migrate] migrate to db version %d, update failed: %s", currentDBVersion+1, err.Error())
			return err
		}
		currentDBVersion++
	}

	return nil
}
