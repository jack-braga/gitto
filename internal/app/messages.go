package app

// Message types are defined in internal/types/messages.go to avoid import cycles.
// This file re-exports them for convenience within the app package.

import "github.com/jack-braga/gitto/internal/types"

// Type aliases for convenience.
type (
	ReposDiscoveredMsg    = types.ReposDiscoveredMsg
	RepoStatusMsg         = types.RepoStatusMsg
	BatchStatusMsg        = types.BatchStatusMsg
	GitOpCompleteMsg      = types.GitOpCompleteMsg
	BranchListMsg         = types.BranchListMsg
	StashListMsg          = types.StashListMsg
	StashDataMsg          = types.StashDataMsg
	LogMsg                = types.LogMsg
	FileTreeMsg           = types.FileTreeMsg
	TickMsg               = types.TickMsg
	StatusNotificationMsg = types.StatusNotificationMsg
	ErrMsg                = types.ErrMsg
	DrillInMsg            = types.DrillInMsg
)
