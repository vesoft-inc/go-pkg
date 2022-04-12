package notify

import (
	"context"
	"crypto"
	"encoding/hex"
	"fmt"
	"sync"
	"time"
)

const (
	DefaultDuplicateFilterInterval   = time.Hour
	DefaultDuplicateFilterMaxRecords = 100
)

var (
	_ Filter   = (*duplicateFilter)(nil)
	_ Filter   = FilterFunc(nil)
	_ Notifier = (*defaultFilterNotifier)(nil)
)

type (
	// Filter is to filter whether a notify should to be sent.
	Filter interface {
		// IfNotify's arguments is the arguments other than the first one in Notifier.Notify.
		IfNotify(interface{}) bool
	}

	// FilterFunc is an adapter to allow the use of ordinary functions as Filter.
	FilterFunc func(interface{}) bool

	defaultFilterNotifier struct {
		filter   Filter
		notifier Notifier
	}

	DuplicateFilterParams struct {
		DupInterval time.Duration // duplicate data notification interval
		MaxRecords  int           // the max of records
	}

	duplicateFilter struct {
		params      DuplicateFilterParams
		mu          sync.RWMutex         // it's for safe access lastTimeMap
		lastTimeMap map[string]time.Time // record the time of last notify, the key is the hash of the data
	}
)

// NewWithFilter creates Notifier for notifiers with a Filter.
func NewWithFilter(f Filter, notifiers ...Notifier) Notifier {
	return &defaultFilterNotifier{
		filter:   f,
		notifier: combineNotifiers(notifiers...),
	}
}

// NewWithDuplicateFilter creates Notifier for notifiers with a Filter.
// Within dupInterval, the same message will only be notify once.
func NewWithDuplicateFilter(params DuplicateFilterParams, notifiers ...Notifier) Notifier {
	return NewWithFilter(newDuplicateFilter(params), notifiers...)
}

func newDuplicateFilter(params DuplicateFilterParams) Filter {
	if params.DupInterval <= 0 {
		params.DupInterval = DefaultDuplicateFilterInterval
	}
	if params.MaxRecords <= 0 {
		params.MaxRecords = DefaultDuplicateFilterMaxRecords
	}
	return &duplicateFilter{
		params:      params,
		lastTimeMap: map[string]time.Time{},
	}
}

func (n *defaultFilterNotifier) Notify(ctx context.Context, data interface{}) error {
	if n.filter != nil && !n.filter.IfNotify(data) {
		return nil
	}
	return n.notifier.Notify(ctx, data)
}

func (n *duplicateFilter) IfNotify(data interface{}) bool {
	hashFn := func(values ...interface{}) string {
		h := crypto.MD5.New()
		for _, v := range values {
			_, _ = fmt.Fprint(h, v)
		}

		return hex.EncodeToString(h.Sum(nil))
	}

	hash := hashFn(data)

	n.mu.RLock()
	lastTime, ok := n.lastTimeMap[hash]
	n.mu.RUnlock()

	if ok && n.isInCoolPeriod(lastTime) {
		return false
	}

	n.mu.Lock()
	n.lastTimeMap[hash] = time.Now()
	n.mu.Unlock()

	n.cleanLastTimeMapIfNecessary()

	return true
}

func (n *duplicateFilter) cleanLastTimeMapIfNecessary() {
	var count int
	n.mu.RLock()
	count = len(n.lastTimeMap)
	n.mu.RUnlock()

	if count < n.params.MaxRecords {
		return
	}

	newMap := make(map[string]time.Time)
	n.mu.Lock()
	for hash, lastTime := range n.lastTimeMap {
		if n.isInCoolPeriod(lastTime) {
			newMap[hash] = lastTime
		}
	}
	n.lastTimeMap = newMap
	n.mu.Unlock()
}

func (n *duplicateFilter) isInCoolPeriod(lastTime time.Time) bool {
	return time.Since(lastTime) <= n.params.DupInterval
}

func (f FilterFunc) IfNotify(data interface{}) bool {
	return f != nil && f(data)
}
