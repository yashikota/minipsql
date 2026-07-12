package minipsql

import "errors"

// ErrEngineUnavailable is returned by development builds that do not yet
// contain the generated pgrust engine.
var ErrEngineUnavailable = errors.New("minipsql: generated pgrust engine is not available")
