package types

import (
	"cosmossdk.io/errors"
)

// Module sentinel errors
var (
	ErrInvalidOwner          = errors.Register(ModuleName, 1, "invalid owner address")
	ErrInvalidAgentKey       = errors.Register(ModuleName, 2, "invalid agent key")
	ErrAgentNotFound         = errors.Register(ModuleName, 3, "agent not found")
	ErrAgentNotActive        = errors.Register(ModuleName, 4, "agent is not active")
	ErrAgentExpired          = errors.Register(ModuleName, 5, "agent subscription expired")
	ErrUnauthorized          = errors.Register(ModuleName, 6, "unauthorized action")
	ErrPermissionDenied      = errors.Register(ModuleName, 7, "permission denied")
	ErrSpendingLimitExceeded = errors.Register(ModuleName, 8, "spending limit exceeded")
	ErrRateLimitExceeded     = errors.Register(ModuleName, 9, "rate limit exceeded")
	ErrInvalidPermission     = errors.Register(ModuleName, 10, "invalid permission")
	ErrInvalidAction         = errors.Register(ModuleName, 11, "invalid action")
	ErrMaxAgentsReached      = errors.Register(ModuleName, 12, "maximum agents reached for subscription tier")
	ErrInvalidSubscription   = errors.Register(ModuleName, 13, "invalid subscription tier")
	ErrModuleNotAllowed      = errors.Register(ModuleName, 14, "module not allowed for subscription tier")
	ErrActionFailed          = errors.Register(ModuleName, 15, "action execution failed")
	ErrInvalidParams         = errors.Register(ModuleName, 16, "invalid action parameters")
	ErrAgentAlreadyExists    = errors.Register(ModuleName, 17, "agent with this key already exists")
)
