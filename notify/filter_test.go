package notify

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestNewWithFilter(t *testing.T) {
	var (
		ast       = assert.New(t)
		notifier  Notifier
		notifiers []Notifier
	)

	notifier = NewWithFilter(FilterFunc(nil), NotifierFunc(nil))
	ast.IsType(&defaultFilterNotifier{}, notifier)
	notifier = notifier.(*defaultFilterNotifier).notifier
	ast.IsType(NotifierFunc(nil), notifier)

	notifier = NewWithFilter(FilterFunc(nil), NotifierFunc(nil), NotifierFunc(nil))
	ast.IsType(&defaultFilterNotifier{}, notifier)
	notifier = notifier.(*defaultFilterNotifier).notifier
	ast.IsType(&defaultNotify{}, notifier)
	notifiers = notifier.(*defaultNotify).notifiers
	ast.Len(notifiers, 2)
}

func TestNewWithDuplicateFilter(t *testing.T) {
	var (
		ast       = assert.New(t)
		notifier  Notifier
		notifiers []Notifier
		filter    Filter
	)

	notifier = NewWithDuplicateFilter(DuplicateFilterParams{}, NotifierFunc(nil))
	ast.IsType(&defaultFilterNotifier{}, notifier)
	filter = notifier.(*defaultFilterNotifier).filter
	ast.IsType(&duplicateFilter{}, filter)
	ast.Equal(DefaultDuplicateFilterInterval, filter.(*duplicateFilter).params.DupInterval)
	ast.Equal(DefaultDuplicateFilterMaxRecords, filter.(*duplicateFilter).params.MaxRecords)
	notifier = notifier.(*defaultFilterNotifier).notifier
	ast.IsType(NotifierFunc(nil), notifier)

	notifier = NewWithDuplicateFilter(DuplicateFilterParams{}, NotifierFunc(nil), NotifierFunc(nil))
	ast.IsType(&defaultFilterNotifier{}, notifier)
	filter = notifier.(*defaultFilterNotifier).filter
	ast.IsType(&duplicateFilter{}, filter)
	ast.Equal(DefaultDuplicateFilterInterval, filter.(*duplicateFilter).params.DupInterval)
	ast.Equal(DefaultDuplicateFilterMaxRecords, filter.(*duplicateFilter).params.MaxRecords)
	notifier = notifier.(*defaultFilterNotifier).notifier
	ast.IsType(&defaultNotify{}, notifier)
	notifiers = notifier.(*defaultNotify).notifiers
	ast.Len(notifiers, 2)
}

func TestNewWithFilterMulti(t *testing.T) {
	var (
		notifier Notifier
		err      error
	)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ast := assert.New(t)

	var f1Count int32
	f1 := FilterFunc(func(interface{}) (fnSubmit func(), needNotify bool) {
		return func() {
			atomic.AddInt32(&f1Count, 1)
		}, false
	})
	var f2Count int32
	f2 := FilterFunc(func(interface{}) (fnSubmit func(), needNotify bool) {
		return func() {
			atomic.AddInt32(&f2Count, 1)
		}, true
	})
	f3 := FilterFunc(nil)

	n11 := NewMockNotifier(ctrl)
	n21 := NewMockNotifier(ctrl)
	n31 := NewMockNotifier(ctrl)

	n1 := NewWithFilter(f1, n11)
	n2 := NewWithFilter(f2, n21)
	n3 := NewWithFilter(f3, n31)
	notifier = New(n1, n2, n3)

	n11.EXPECT().Notify(gomock.Any(), gomock.Any()).Times(0)
	n21.EXPECT().Notify(gomock.Any(), gomock.Any()).Times(1)
	n31.EXPECT().Notify(gomock.Any(), gomock.Any()).Times(1)

	err = notifier.Notify(context.TODO(), "Message")
	ast.NoError(err)
	ast.Equal(f1Count, int32(0))
	ast.Equal(f2Count, int32(1))
}

func TestNewWithFilterFailed(t *testing.T) {
	var (
		notifier Notifier
		err      error
	)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ast := assert.New(t)

	var fCount int32
	f := FilterFunc(func(interface{}) (fnSubmit func(), needNotify bool) {
		return func() {
			atomic.AddInt32(&fCount, 1)
		}, true
	})

	n11 := NewMockNotifier(ctrl)

	n1 := NewWithFilter(f, n11)
	notifier = New(n1)

	var notifyCount = 10
	n11.EXPECT().Notify(gomock.Any(), gomock.Any()).Times(notifyCount).Return(errors.New("sendError"))
	for i := 0; i < notifyCount; i++ {
		err = notifier.Notify(context.TODO(), "Message")
		if ast.ErrorIs(err, ErrNotifyNotification) {
			ast.Contains(err.Error(), "sendError")
		}
	}
	ast.Equal(fCount, int32(0))
}

func TestDuplicateFilterNotify(t *testing.T) {
	var (
		notifier    Notifier
		err         error
		dupInterval = 10 * time.Millisecond
	)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ast := assert.New(t)

	n11 := NewMockNotifier(ctrl)
	n12 := NewMockNotifier(ctrl)
	n13 := NewMockNotifier(ctrl)
	n21 := NewMockNotifier(ctrl)
	n22 := NewMockNotifier(ctrl)
	n23 := NewMockNotifier(ctrl)

	n1 := New(
		NewWithDuplicateFilter(DuplicateFilterParams{
			DupInterval: dupInterval,
			MaxRecords:  1,
		}, n11, n12),
		n13,
	)
	n2 := New(
		NewWithDuplicateFilter(DuplicateFilterParams{
			DupInterval: dupInterval,
			MaxRecords:  10,
		}, n21),
		n22,
		n23,
	)

	notifier = New(n1, n2)

	n11.EXPECT().Notify(gomock.Any(), gomock.Any()).Times(1)
	n12.EXPECT().Notify(gomock.Any(), gomock.Any()).Times(1)
	n13.EXPECT().Notify(gomock.Any(), gomock.Any()).Times(1)
	n21.EXPECT().Notify(gomock.Any(), gomock.Any()).Times(1)
	n22.EXPECT().Notify(gomock.Any(), gomock.Any()).Times(1)
	n23.EXPECT().Notify(gomock.Any(), gomock.Any()).Times(1)
	err = notifier.Notify(context.TODO(), "Message")
	ast.NoError(err)
	if err != nil {
		panic(err)
	}

	n13.EXPECT().Notify(gomock.Any(), gomock.Any()).Times(1)
	n22.EXPECT().Notify(gomock.Any(), gomock.Any()).Times(1)
	n23.EXPECT().Notify(gomock.Any(), gomock.Any()).Times(1)
	err = notifier.Notify(context.TODO(), "Message")
	ast.NoError(err)

	n11.EXPECT().Notify(gomock.Any(), gomock.Any()).Times(1)
	n12.EXPECT().Notify(gomock.Any(), gomock.Any()).Times(1)
	n13.EXPECT().Notify(gomock.Any(), gomock.Any()).Times(1)
	n21.EXPECT().Notify(gomock.Any(), gomock.Any()).Times(1)
	n22.EXPECT().Notify(gomock.Any(), gomock.Any()).Times(1)
	n23.EXPECT().Notify(gomock.Any(), gomock.Any()).Times(1)
	err = notifier.Notify(context.TODO(), map[string]interface{}{
		"i":  100,
		"f":  9.2,
		"s":  "str",
		"bs": []byte("bytes"),
	})
	ast.NoError(err)

	n13.EXPECT().Notify(gomock.Any(), gomock.Any()).Times(1)
	n22.EXPECT().Notify(gomock.Any(), gomock.Any()).Times(1)
	n23.EXPECT().Notify(gomock.Any(), gomock.Any()).Times(1)
	err = notifier.Notify(context.TODO(), map[string]interface{}{
		"i":  100,
		"f":  9.2,
		"s":  "str",
		"bs": []byte("bytes"),
	})
	ast.NoError(err)

	time.Sleep(dupInterval)

	n11.EXPECT().Notify(gomock.Any(), gomock.Any()).Times(1)
	n12.EXPECT().Notify(gomock.Any(), gomock.Any()).Times(1)
	n13.EXPECT().Notify(gomock.Any(), gomock.Any()).Times(1)
	n21.EXPECT().Notify(gomock.Any(), gomock.Any()).Times(1)
	n22.EXPECT().Notify(gomock.Any(), gomock.Any()).Times(1)
	n23.EXPECT().Notify(gomock.Any(), gomock.Any()).Times(1)
	err = notifier.Notify(context.TODO(), "Message")
	ast.NoError(err)

	n13.EXPECT().Notify(gomock.Any(), gomock.Any()).Times(1)
	n22.EXPECT().Notify(gomock.Any(), gomock.Any()).Times(1)
	n23.EXPECT().Notify(gomock.Any(), gomock.Any()).Times(1)
	err = notifier.Notify(context.TODO(), "Message")
	ast.NoError(err)

	var findFilter func(n Notifier) *duplicateFilter
	findFilter = func(n Notifier) *duplicateFilter {
		switch v := n.(type) {
		case *defaultNotify:
			for i := range v.notifiers {
				if df := findFilter(v.notifiers[i]); df != nil {
					return df
				}
			}
		case *defaultFilterNotifier:
			return v.filter.(*duplicateFilter)
		}
		return nil
	}

	if df := findFilter(n1); ast.NotNil(df) {
		ast.Len(df.lastTimeMap, 1)
	}
	if df := findFilter(n2); ast.NotNil(df) {
		ast.Len(df.lastTimeMap, 2)
	}
}
