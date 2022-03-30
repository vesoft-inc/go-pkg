package notify

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	var (
		ast      = assert.New(t)
		notifier Notifier
	)

	notifier = New()
	ast.IsType(&defaultNotify{}, notifier)

	var (
		fnLength = 6
		intMap   = make(map[int]struct{})
		mu       sync.Mutex
	)

	appendInt := func(i int) {
		mu.Lock()
		intMap[i] = struct{}{}
		mu.Unlock()
	}

	fn0 := func(context.Context, interface{}) error { appendInt(0); return nil }
	fn1 := func(context.Context, interface{}) error { appendInt(1); return nil }
	fn2 := func(context.Context, interface{}) error { appendInt(2); return nil }
	fn3 := func(context.Context, string) error { appendInt(3); return nil }
	fn4 := func(context.Context, string) error { appendInt(4); return nil }
	fn5 := func(context.Context, interface{}) error { appendInt(5); return nil }

	notifier = New(
		NotifierFunc(fn0), NotifierFunc(fn1), NotifierFunc(fn2),
		stringNotifierToNotifier(StringNotifierFunc(fn3)), StringNotifierFunc(fn4).Notifier(),
		NotifierFunc(fn5),
	)

	ast.IsType(&defaultNotify{}, notifier)

	notifiers := notifier.(*defaultNotify).notifiers
	ast.Len(notifiers, fnLength)

	err := notifier.Notify(context.TODO(), nil)
	ast.NoError(err)
	ast.Equal(map[int]struct{}{0: {}, 1: {}, 2: {}, 3: {}, 4: {}, 5: {}}, intMap)
}

func TestNotify(t *testing.T) {
	var (
		notifier Notifier
		results  = make(map[string]interface{})
		mu       sync.Mutex
		err      error
	)

	ast := assert.New(t)

	cleanResults := func() {
		mu.Lock()
		defer mu.Unlock()
		results = make(map[string]interface{})
	}
	addResults := func(subject string, data interface{}) {
		mu.Lock()
		defer mu.Unlock()
		results[subject] = data
	}

	notifiers := []Notifier{
		NewWithStringNotifiers(
			StringNotifierFunc(func(_ context.Context, message string) error {
				addResults("a", message)
				return nil
			}),
		),
		StringNotifierFunc(nil).Notifier(),
		NotifierFunc(func(_ context.Context, data interface{}) error {
			addResults("b", data)
			return nil
		}),
		NotifierFunc(nil),
		NotifierFunc(func(_ context.Context, data interface{}) error {
			addResults("c", data)
			return nil
		}),
	}

	notifier = New(notifiers...)

	for _, data := range []interface{}{
		100,
		9.2,
		"str",
		[]byte("bytes"),
		[]interface{}{100, 9.2, "str", []byte("bytes")},
		map[string]interface{}{
			"i":  100,
			"f":  9.2,
			"s":  "str",
			"bs": []byte("bytes"),
		},
	} {
		cleanResults()
		err = notifier.Notify(context.TODO(), data)
		ast.NoError(err)
		ast.Len(results, 3)
		ast.Equal(map[string]interface{}{
			"a": convertDataMessage(data),
			"b": data,
			"c": data,
		}, results)
	}

	expectedError := errors.New("testError")
	notifier = New(append(notifiers, NotifierFunc(func(_ context.Context, data interface{}) error {
		return expectedError
	}))...)
	err = notifier.Notify(context.TODO(), "data")
	if ast.ErrorIs(err, ErrNotifyNotification) {
		ast.Contains(err.Error(), expectedError.Error())
	}
}

func Test_combineNotifiers(t *testing.T) {
	var (
		ast       = assert.New(t)
		notifier  Notifier
		notifiers []Notifier
		err       error
	)

	notifier = combineNotifiers(notifiers...)
	ast.IsType(NotifierFunc(nil), notifier)
	ast.Nil(notifier.(NotifierFunc))
	err = notifier.Notify(context.TODO(), "")
	ast.NoError(err)

	notifiers = []Notifier{NotifierFunc(func(_ context.Context, data interface{}) error { return nil })}
	notifier = combineNotifiers(notifiers...)
	ast.IsType(NotifierFunc(nil), notifier)
	ast.NotNil(notifier.(NotifierFunc))
	err = notifier.Notify(context.TODO(), "")
	ast.NoError(err)

	notifiers = []Notifier{
		NotifierFunc(func(_ context.Context, data interface{}) error { return nil }),
		NotifierFunc(func(_ context.Context, data interface{}) error { return nil }),
	}
	notifier = combineNotifiers(notifiers...)
	ast.IsType(&defaultNotify{}, notifier)
	ast.Len(notifier.(*defaultNotify).notifiers, 2)
	err = notifier.Notify(context.TODO(), "")
	ast.NoError(err)
}
