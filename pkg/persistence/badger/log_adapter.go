package badger

import "go.uber.org/zap"

type badgerLogAdapter struct {
	*zap.SugaredLogger
}

func (b *badgerLogAdapter) Warningf(s string, i ...interface{}) {
	b.Warnf(s, i...)
}
