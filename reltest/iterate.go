package reltest

import (
	"context"
	"fmt"
	"io"
	"reflect"

	"github.com/go-rel/rel"
)

type iterate []*MockIterate

func (i *iterate) register(ctxData ctxData, query rel.Query, options ...rel.IteratorOption) *MockIterate {
	mi := &MockIterate{
		assert:     &Assert{ctxData: ctxData},
		argQuery:   query,
		argOptions: options,
	}
	*i = append(*i, mi)
	return mi
}

func (i iterate) execute(ctx context.Context, query rel.Query, options ...rel.IteratorOption) rel.Iterator {
	for _, mi := range i {
		if reflect.DeepEqual(mi.argOptions, options) &&
			matchQuery(mi.argQuery, query) &&
			mi.assert.call(ctx) {
			return mi
		}
	}

	mi := MockIterate{argQuery: query, argOptions: options}
	mocks := ""
	for x := range i {
		mocks += "\n\t" + i[x].ExpectString()
	}
	panic(fmt.Sprintf("FAIL: this call is not mocked:\n\t%s\nMaybe try adding mock:\t\n%s\n\nAvailable mocks:%s", mi, mi.ExpectString(), mocks))
}

type data interface {
	Len() int
	Get(index int) *rel.Document
}

// MockIterate asserts and simulate Delete function for test.
type MockIterate struct {
	assert     *Assert
	result     data
	current    int
	err        error
	argQuery   rel.Query
	argOptions []rel.IteratorOption
}

// Result sets the result of preload.
func (mi *MockIterate) Result(result interface{}) *Assert {
	rt := reflect.TypeOf(result)
	if rt.Kind() == reflect.Ptr {
		rt = rt.Elem()
	}

	if rt.Kind() == reflect.Slice {
		mi.result = rel.NewCollection(result, true)
	} else {
		mi.result = rel.NewDocument(result, true)
	}
	return mi.assert
}

// Error sets error to be returned.
func (mi *MockIterate) Error(err error) *Assert {
	mi.err = err
	return mi.assert
}

// ConnectionClosed sets this error to be returned.
func (mi *MockIterate) ConnectionClosed() *Assert {
	return mi.Error(ErrConnectionClosed)
}

func (mi MockIterate) Close() error {
	return nil
}

func (mi *MockIterate) Next(record interface{}) error {
	if mi.err != nil {
		return mi.err
	}

	if mi.result == nil || mi.current == mi.result.Len() {
		return io.EOF
	}

	var (
		doc = mi.result.Get(mi.current)
	)

	reflect.ValueOf(record).Elem().Set(doc.ReflectValue())

	mi.current++
	return nil
}

// String representation of mocked call.
func (mi MockIterate) String() string {
	argOptions := ""
	for i := range mi.argOptions {
		argOptions += fmt.Sprintf(", %v", mi.argOptions[i])
	}

	return fmt.Sprintf("Iterate(ctx, %s%s)", mi.argQuery, argOptions)
}

// ExpectString representation of mocked call.
func (mi MockIterate) ExpectString() string {
	argOptions := ""
	for i := range mi.argOptions {
		argOptions += fmt.Sprintf(", %v", mi.argOptions[i])
	}

	return fmt.Sprintf("ExpectIterate(ctx, %s%s)", mi.argQuery, argOptions)
}
