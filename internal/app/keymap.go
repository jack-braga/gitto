package app

import "github.com/jack-braga/gitto/internal/types"

// KeyMap is re-exported from the types package.
type KeyMap = types.KeyMap

// DefaultKeyMap returns the default keybindings.
func DefaultKeyMap() KeyMap {
	return types.DefaultKeyMap()
}
