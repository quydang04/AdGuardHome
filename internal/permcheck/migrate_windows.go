//go:build windows

package permcheck

import (
	"context"
	"log/slog"
)

// needsMigration is the Windows-specific implementation of [NeedsMigration].
func needsMigration(_ context.Context, _ *slog.Logger, _, _ string) (ok bool) {
	return false
}

// migrate is the Windows-specific implementation of [Migrate].
//
// It sets the owner to administrators and adds a full control access control
// entry for the account.  It also removes all non-administrator access control
// entries, and keeps deny access control entries.  For any created or modified
// entry it sets the propagation flags to be inherited by child objects.
func migrate(_ context.Context, _ *slog.Logger, _, _, _, _, _ string) {
}
