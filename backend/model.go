package backend

import (
	"time"

	"golang.org/x/net/context"

	"google.golang.org/appengine/datastore"
)

// EchoModel is an object with data to return in an echo request
type EchoModel struct {
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CurrentDate time.Time `json:"currentDate"`
}

// Counter is just a counter
type Counter struct {
	Count int64
}

// Gets a key for a specified counter by name
func counterKey(c context.Context, counterName string) *datastore.Key {
	return datastore.NewKey(c, "Counter", counterName, 0, nil)
}

// NumberOf Gets the number of entities
func NumberOf(c context.Context, counterName string) (int64, error) {
	k := counterKey(c, counterName)
	var counter Counter
	if err := datastore.Get(c, k, &counter); err != nil {
		return 0, err
	}
	return counter.Count, nil
}

func updateCounter(c context.Context, counterName string) error {
	key := counterKey(c, counterName)
	count := new(Counter)
	err := datastore.RunInTransaction(c, func(c context.Context) error {
		err := datastore.Get(c, key, count)
		if err != nil && err != datastore.ErrNoSuchEntity {
			return err
		}
		count.Count++
		_, err = datastore.Put(c, key, count)
		return err
	}, nil)

	return err
}
