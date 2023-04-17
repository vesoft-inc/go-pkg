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
	// Filter is to filter whether notify should to be sent.
	Filter interface {
		// IfNotify check whether notify should to be sent.
		// The arguments are the arguments other than the first one in Notifier.Notify.
		IfNotify(interface{}) (fnSubmit func(), needNotify bool)
	}

	// FilterFunc is an adapter to allow the use of ordinary functions as Filter.
	FilterFunc func(data interface{}) (fnSubmit func(), needNotify bool)

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
	var (
		fnSubmit   func()
		needNotify = true
	)
	if n.filter != nil {
		fnSubmit, needNotify = n.filter.IfNotify(data)
	}

	if !needNotify {
		return nil
	}
	if err := n.notifier.Notify(ctx, data); err != nil {
		return err
	}

	// submit only if successfully
	if fnSubmit != nil {
		fnSubmit()
	}

	return nil
}

func (n *duplicateFilter) IfNotify(data interface{}) (fnSubmit func(), needNotify bool) {
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
		return nil, false
	}

	n.cleanLastTimeMapIfNecessary()

	return func() {
		n.mu.Lock()
		n.lastTimeMap[hash] = time.Now()
		n.mu.Unlock()
	}, true
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

func (f FilterFunc) IfNotify(data interface{}) (fnSubmit func(), needNotify bool) {
	if f == nil { // default to notify
		return func() {}, true
	}
	return f(data)
}
