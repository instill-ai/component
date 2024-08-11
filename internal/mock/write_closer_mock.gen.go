// Code generated by http://github.com/gojuno/minimock (v3.3.11). DO NOT EDIT.

package mock

import (
	"sync"
	mm_atomic "sync/atomic"
	mm_time "time"

	"github.com/gojuno/minimock/v3"
)

// WriteCloserMock implements io.WriteCloser
type WriteCloserMock struct {
	t          minimock.Tester
	finishOnce sync.Once

	funcClose          func() (err error)
	inspectFuncClose   func()
	afterCloseCounter  uint64
	beforeCloseCounter uint64
	CloseMock          mWriteCloserMockClose

	funcWrite          func(p []byte) (n int, err error)
	inspectFuncWrite   func(p []byte)
	afterWriteCounter  uint64
	beforeWriteCounter uint64
	WriteMock          mWriteCloserMockWrite
}

// NewWriteCloserMock returns a mock for io.WriteCloser
func NewWriteCloserMock(t minimock.Tester) *WriteCloserMock {
	m := &WriteCloserMock{t: t}

	if controller, ok := t.(minimock.MockController); ok {
		controller.RegisterMocker(m)
	}

	m.CloseMock = mWriteCloserMockClose{mock: m}

	m.WriteMock = mWriteCloserMockWrite{mock: m}
	m.WriteMock.callArgs = []*WriteCloserMockWriteParams{}

	t.Cleanup(m.MinimockFinish)

	return m
}

type mWriteCloserMockClose struct {
	optional           bool
	mock               *WriteCloserMock
	defaultExpectation *WriteCloserMockCloseExpectation
	expectations       []*WriteCloserMockCloseExpectation

	expectedInvocations uint64
}

// WriteCloserMockCloseExpectation specifies expectation struct of the WriteCloser.Close
type WriteCloserMockCloseExpectation struct {
	mock *WriteCloserMock

	results *WriteCloserMockCloseResults
	Counter uint64
}

// WriteCloserMockCloseResults contains results of the WriteCloser.Close
type WriteCloserMockCloseResults struct {
	err error
}

// Marks this method to be optional. The default behavior of any method with Return() is '1 or more', meaning
// the test will fail minimock's automatic final call check if the mocked method was not called at least once.
// Optional() makes method check to work in '0 or more' mode.
// It is NOT RECOMMENDED to use this option by default unless you really need it, as it helps to
// catch the problems when the expected method call is totally skipped during test run.
func (mmClose *mWriteCloserMockClose) Optional() *mWriteCloserMockClose {
	mmClose.optional = true
	return mmClose
}

// Expect sets up expected params for WriteCloser.Close
func (mmClose *mWriteCloserMockClose) Expect() *mWriteCloserMockClose {
	if mmClose.mock.funcClose != nil {
		mmClose.mock.t.Fatalf("WriteCloserMock.Close mock is already set by Set")
	}

	if mmClose.defaultExpectation == nil {
		mmClose.defaultExpectation = &WriteCloserMockCloseExpectation{}
	}

	return mmClose
}

// Inspect accepts an inspector function that has same arguments as the WriteCloser.Close
func (mmClose *mWriteCloserMockClose) Inspect(f func()) *mWriteCloserMockClose {
	if mmClose.mock.inspectFuncClose != nil {
		mmClose.mock.t.Fatalf("Inspect function is already set for WriteCloserMock.Close")
	}

	mmClose.mock.inspectFuncClose = f

	return mmClose
}

// Return sets up results that will be returned by WriteCloser.Close
func (mmClose *mWriteCloserMockClose) Return(err error) *WriteCloserMock {
	if mmClose.mock.funcClose != nil {
		mmClose.mock.t.Fatalf("WriteCloserMock.Close mock is already set by Set")
	}

	if mmClose.defaultExpectation == nil {
		mmClose.defaultExpectation = &WriteCloserMockCloseExpectation{mock: mmClose.mock}
	}
	mmClose.defaultExpectation.results = &WriteCloserMockCloseResults{err}
	return mmClose.mock
}

// Set uses given function f to mock the WriteCloser.Close method
func (mmClose *mWriteCloserMockClose) Set(f func() (err error)) *WriteCloserMock {
	if mmClose.defaultExpectation != nil {
		mmClose.mock.t.Fatalf("Default expectation is already set for the WriteCloser.Close method")
	}

	if len(mmClose.expectations) > 0 {
		mmClose.mock.t.Fatalf("Some expectations are already set for the WriteCloser.Close method")
	}

	mmClose.mock.funcClose = f
	return mmClose.mock
}

// Times sets number of times WriteCloser.Close should be invoked
func (mmClose *mWriteCloserMockClose) Times(n uint64) *mWriteCloserMockClose {
	if n == 0 {
		mmClose.mock.t.Fatalf("Times of WriteCloserMock.Close mock can not be zero")
	}
	mm_atomic.StoreUint64(&mmClose.expectedInvocations, n)
	return mmClose
}

func (mmClose *mWriteCloserMockClose) invocationsDone() bool {
	if len(mmClose.expectations) == 0 && mmClose.defaultExpectation == nil && mmClose.mock.funcClose == nil {
		return true
	}

	totalInvocations := mm_atomic.LoadUint64(&mmClose.mock.afterCloseCounter)
	expectedInvocations := mm_atomic.LoadUint64(&mmClose.expectedInvocations)

	return totalInvocations > 0 && (expectedInvocations == 0 || expectedInvocations == totalInvocations)
}

// Close implements io.WriteCloser
func (mmClose *WriteCloserMock) Close() (err error) {
	mm_atomic.AddUint64(&mmClose.beforeCloseCounter, 1)
	defer mm_atomic.AddUint64(&mmClose.afterCloseCounter, 1)

	if mmClose.inspectFuncClose != nil {
		mmClose.inspectFuncClose()
	}

	if mmClose.CloseMock.defaultExpectation != nil {
		mm_atomic.AddUint64(&mmClose.CloseMock.defaultExpectation.Counter, 1)

		mm_results := mmClose.CloseMock.defaultExpectation.results
		if mm_results == nil {
			mmClose.t.Fatal("No results are set for the WriteCloserMock.Close")
		}
		return (*mm_results).err
	}
	if mmClose.funcClose != nil {
		return mmClose.funcClose()
	}
	mmClose.t.Fatalf("Unexpected call to WriteCloserMock.Close.")
	return
}

// CloseAfterCounter returns a count of finished WriteCloserMock.Close invocations
func (mmClose *WriteCloserMock) CloseAfterCounter() uint64 {
	return mm_atomic.LoadUint64(&mmClose.afterCloseCounter)
}

// CloseBeforeCounter returns a count of WriteCloserMock.Close invocations
func (mmClose *WriteCloserMock) CloseBeforeCounter() uint64 {
	return mm_atomic.LoadUint64(&mmClose.beforeCloseCounter)
}

// MinimockCloseDone returns true if the count of the Close invocations corresponds
// the number of defined expectations
func (m *WriteCloserMock) MinimockCloseDone() bool {
	if m.CloseMock.optional {
		// Optional methods provide '0 or more' call count restriction.
		return true
	}

	for _, e := range m.CloseMock.expectations {
		if mm_atomic.LoadUint64(&e.Counter) < 1 {
			return false
		}
	}

	return m.CloseMock.invocationsDone()
}

// MinimockCloseInspect logs each unmet expectation
func (m *WriteCloserMock) MinimockCloseInspect() {
	for _, e := range m.CloseMock.expectations {
		if mm_atomic.LoadUint64(&e.Counter) < 1 {
			m.t.Error("Expected call to WriteCloserMock.Close")
		}
	}

	afterCloseCounter := mm_atomic.LoadUint64(&m.afterCloseCounter)
	// if default expectation was set then invocations count should be greater than zero
	if m.CloseMock.defaultExpectation != nil && afterCloseCounter < 1 {
		m.t.Error("Expected call to WriteCloserMock.Close")
	}
	// if func was set then invocations count should be greater than zero
	if m.funcClose != nil && afterCloseCounter < 1 {
		m.t.Error("Expected call to WriteCloserMock.Close")
	}

	if !m.CloseMock.invocationsDone() && afterCloseCounter > 0 {
		m.t.Errorf("Expected %d calls to WriteCloserMock.Close but found %d calls",
			mm_atomic.LoadUint64(&m.CloseMock.expectedInvocations), afterCloseCounter)
	}
}

type mWriteCloserMockWrite struct {
	optional           bool
	mock               *WriteCloserMock
	defaultExpectation *WriteCloserMockWriteExpectation
	expectations       []*WriteCloserMockWriteExpectation

	callArgs []*WriteCloserMockWriteParams
	mutex    sync.RWMutex

	expectedInvocations uint64
}

// WriteCloserMockWriteExpectation specifies expectation struct of the WriteCloser.Write
type WriteCloserMockWriteExpectation struct {
	mock      *WriteCloserMock
	params    *WriteCloserMockWriteParams
	paramPtrs *WriteCloserMockWriteParamPtrs
	results   *WriteCloserMockWriteResults
	Counter   uint64
}

// WriteCloserMockWriteParams contains parameters of the WriteCloser.Write
type WriteCloserMockWriteParams struct {
	p []byte
}

// WriteCloserMockWriteParamPtrs contains pointers to parameters of the WriteCloser.Write
type WriteCloserMockWriteParamPtrs struct {
	p *[]byte
}

// WriteCloserMockWriteResults contains results of the WriteCloser.Write
type WriteCloserMockWriteResults struct {
	n   int
	err error
}

// Marks this method to be optional. The default behavior of any method with Return() is '1 or more', meaning
// the test will fail minimock's automatic final call check if the mocked method was not called at least once.
// Optional() makes method check to work in '0 or more' mode.
// It is NOT RECOMMENDED to use this option by default unless you really need it, as it helps to
// catch the problems when the expected method call is totally skipped during test run.
func (mmWrite *mWriteCloserMockWrite) Optional() *mWriteCloserMockWrite {
	mmWrite.optional = true
	return mmWrite
}

// Expect sets up expected params for WriteCloser.Write
func (mmWrite *mWriteCloserMockWrite) Expect(p []byte) *mWriteCloserMockWrite {
	if mmWrite.mock.funcWrite != nil {
		mmWrite.mock.t.Fatalf("WriteCloserMock.Write mock is already set by Set")
	}

	if mmWrite.defaultExpectation == nil {
		mmWrite.defaultExpectation = &WriteCloserMockWriteExpectation{}
	}

	if mmWrite.defaultExpectation.paramPtrs != nil {
		mmWrite.mock.t.Fatalf("WriteCloserMock.Write mock is already set by ExpectParams functions")
	}

	mmWrite.defaultExpectation.params = &WriteCloserMockWriteParams{p}
	for _, e := range mmWrite.expectations {
		if minimock.Equal(e.params, mmWrite.defaultExpectation.params) {
			mmWrite.mock.t.Fatalf("Expectation set by When has same params: %#v", *mmWrite.defaultExpectation.params)
		}
	}

	return mmWrite
}

// ExpectPParam1 sets up expected param p for WriteCloser.Write
func (mmWrite *mWriteCloserMockWrite) ExpectPParam1(p []byte) *mWriteCloserMockWrite {
	if mmWrite.mock.funcWrite != nil {
		mmWrite.mock.t.Fatalf("WriteCloserMock.Write mock is already set by Set")
	}

	if mmWrite.defaultExpectation == nil {
		mmWrite.defaultExpectation = &WriteCloserMockWriteExpectation{}
	}

	if mmWrite.defaultExpectation.params != nil {
		mmWrite.mock.t.Fatalf("WriteCloserMock.Write mock is already set by Expect")
	}

	if mmWrite.defaultExpectation.paramPtrs == nil {
		mmWrite.defaultExpectation.paramPtrs = &WriteCloserMockWriteParamPtrs{}
	}
	mmWrite.defaultExpectation.paramPtrs.p = &p

	return mmWrite
}

// Inspect accepts an inspector function that has same arguments as the WriteCloser.Write
func (mmWrite *mWriteCloserMockWrite) Inspect(f func(p []byte)) *mWriteCloserMockWrite {
	if mmWrite.mock.inspectFuncWrite != nil {
		mmWrite.mock.t.Fatalf("Inspect function is already set for WriteCloserMock.Write")
	}

	mmWrite.mock.inspectFuncWrite = f

	return mmWrite
}

// Return sets up results that will be returned by WriteCloser.Write
func (mmWrite *mWriteCloserMockWrite) Return(n int, err error) *WriteCloserMock {
	if mmWrite.mock.funcWrite != nil {
		mmWrite.mock.t.Fatalf("WriteCloserMock.Write mock is already set by Set")
	}

	if mmWrite.defaultExpectation == nil {
		mmWrite.defaultExpectation = &WriteCloserMockWriteExpectation{mock: mmWrite.mock}
	}
	mmWrite.defaultExpectation.results = &WriteCloserMockWriteResults{n, err}
	return mmWrite.mock
}

// Set uses given function f to mock the WriteCloser.Write method
func (mmWrite *mWriteCloserMockWrite) Set(f func(p []byte) (n int, err error)) *WriteCloserMock {
	if mmWrite.defaultExpectation != nil {
		mmWrite.mock.t.Fatalf("Default expectation is already set for the WriteCloser.Write method")
	}

	if len(mmWrite.expectations) > 0 {
		mmWrite.mock.t.Fatalf("Some expectations are already set for the WriteCloser.Write method")
	}

	mmWrite.mock.funcWrite = f
	return mmWrite.mock
}

// When sets expectation for the WriteCloser.Write which will trigger the result defined by the following
// Then helper
func (mmWrite *mWriteCloserMockWrite) When(p []byte) *WriteCloserMockWriteExpectation {
	if mmWrite.mock.funcWrite != nil {
		mmWrite.mock.t.Fatalf("WriteCloserMock.Write mock is already set by Set")
	}

	expectation := &WriteCloserMockWriteExpectation{
		mock:   mmWrite.mock,
		params: &WriteCloserMockWriteParams{p},
	}
	mmWrite.expectations = append(mmWrite.expectations, expectation)
	return expectation
}

// Then sets up WriteCloser.Write return parameters for the expectation previously defined by the When method
func (e *WriteCloserMockWriteExpectation) Then(n int, err error) *WriteCloserMock {
	e.results = &WriteCloserMockWriteResults{n, err}
	return e.mock
}

// Times sets number of times WriteCloser.Write should be invoked
func (mmWrite *mWriteCloserMockWrite) Times(n uint64) *mWriteCloserMockWrite {
	if n == 0 {
		mmWrite.mock.t.Fatalf("Times of WriteCloserMock.Write mock can not be zero")
	}
	mm_atomic.StoreUint64(&mmWrite.expectedInvocations, n)
	return mmWrite
}

func (mmWrite *mWriteCloserMockWrite) invocationsDone() bool {
	if len(mmWrite.expectations) == 0 && mmWrite.defaultExpectation == nil && mmWrite.mock.funcWrite == nil {
		return true
	}

	totalInvocations := mm_atomic.LoadUint64(&mmWrite.mock.afterWriteCounter)
	expectedInvocations := mm_atomic.LoadUint64(&mmWrite.expectedInvocations)

	return totalInvocations > 0 && (expectedInvocations == 0 || expectedInvocations == totalInvocations)
}

// Write implements io.WriteCloser
func (mmWrite *WriteCloserMock) Write(p []byte) (n int, err error) {
	mm_atomic.AddUint64(&mmWrite.beforeWriteCounter, 1)
	defer mm_atomic.AddUint64(&mmWrite.afterWriteCounter, 1)

	if mmWrite.inspectFuncWrite != nil {
		mmWrite.inspectFuncWrite(p)
	}

	mm_params := WriteCloserMockWriteParams{p}

	// Record call args
	mmWrite.WriteMock.mutex.Lock()
	mmWrite.WriteMock.callArgs = append(mmWrite.WriteMock.callArgs, &mm_params)
	mmWrite.WriteMock.mutex.Unlock()

	for _, e := range mmWrite.WriteMock.expectations {
		if minimock.Equal(*e.params, mm_params) {
			mm_atomic.AddUint64(&e.Counter, 1)
			return e.results.n, e.results.err
		}
	}

	if mmWrite.WriteMock.defaultExpectation != nil {
		mm_atomic.AddUint64(&mmWrite.WriteMock.defaultExpectation.Counter, 1)
		mm_want := mmWrite.WriteMock.defaultExpectation.params
		mm_want_ptrs := mmWrite.WriteMock.defaultExpectation.paramPtrs

		mm_got := WriteCloserMockWriteParams{p}

		if mm_want_ptrs != nil {

			if mm_want_ptrs.p != nil && !minimock.Equal(*mm_want_ptrs.p, mm_got.p) {
				mmWrite.t.Errorf("WriteCloserMock.Write got unexpected parameter p, want: %#v, got: %#v%s\n", *mm_want_ptrs.p, mm_got.p, minimock.Diff(*mm_want_ptrs.p, mm_got.p))
			}

		} else if mm_want != nil && !minimock.Equal(*mm_want, mm_got) {
			mmWrite.t.Errorf("WriteCloserMock.Write got unexpected parameters, want: %#v, got: %#v%s\n", *mm_want, mm_got, minimock.Diff(*mm_want, mm_got))
		}

		mm_results := mmWrite.WriteMock.defaultExpectation.results
		if mm_results == nil {
			mmWrite.t.Fatal("No results are set for the WriteCloserMock.Write")
		}
		return (*mm_results).n, (*mm_results).err
	}
	if mmWrite.funcWrite != nil {
		return mmWrite.funcWrite(p)
	}
	mmWrite.t.Fatalf("Unexpected call to WriteCloserMock.Write. %v", p)
	return
}

// WriteAfterCounter returns a count of finished WriteCloserMock.Write invocations
func (mmWrite *WriteCloserMock) WriteAfterCounter() uint64 {
	return mm_atomic.LoadUint64(&mmWrite.afterWriteCounter)
}

// WriteBeforeCounter returns a count of WriteCloserMock.Write invocations
func (mmWrite *WriteCloserMock) WriteBeforeCounter() uint64 {
	return mm_atomic.LoadUint64(&mmWrite.beforeWriteCounter)
}

// Calls returns a list of arguments used in each call to WriteCloserMock.Write.
// The list is in the same order as the calls were made (i.e. recent calls have a higher index)
func (mmWrite *mWriteCloserMockWrite) Calls() []*WriteCloserMockWriteParams {
	mmWrite.mutex.RLock()

	argCopy := make([]*WriteCloserMockWriteParams, len(mmWrite.callArgs))
	copy(argCopy, mmWrite.callArgs)

	mmWrite.mutex.RUnlock()

	return argCopy
}

// MinimockWriteDone returns true if the count of the Write invocations corresponds
// the number of defined expectations
func (m *WriteCloserMock) MinimockWriteDone() bool {
	if m.WriteMock.optional {
		// Optional methods provide '0 or more' call count restriction.
		return true
	}

	for _, e := range m.WriteMock.expectations {
		if mm_atomic.LoadUint64(&e.Counter) < 1 {
			return false
		}
	}

	return m.WriteMock.invocationsDone()
}

// MinimockWriteInspect logs each unmet expectation
func (m *WriteCloserMock) MinimockWriteInspect() {
	for _, e := range m.WriteMock.expectations {
		if mm_atomic.LoadUint64(&e.Counter) < 1 {
			m.t.Errorf("Expected call to WriteCloserMock.Write with params: %#v", *e.params)
		}
	}

	afterWriteCounter := mm_atomic.LoadUint64(&m.afterWriteCounter)
	// if default expectation was set then invocations count should be greater than zero
	if m.WriteMock.defaultExpectation != nil && afterWriteCounter < 1 {
		if m.WriteMock.defaultExpectation.params == nil {
			m.t.Error("Expected call to WriteCloserMock.Write")
		} else {
			m.t.Errorf("Expected call to WriteCloserMock.Write with params: %#v", *m.WriteMock.defaultExpectation.params)
		}
	}
	// if func was set then invocations count should be greater than zero
	if m.funcWrite != nil && afterWriteCounter < 1 {
		m.t.Error("Expected call to WriteCloserMock.Write")
	}

	if !m.WriteMock.invocationsDone() && afterWriteCounter > 0 {
		m.t.Errorf("Expected %d calls to WriteCloserMock.Write but found %d calls",
			mm_atomic.LoadUint64(&m.WriteMock.expectedInvocations), afterWriteCounter)
	}
}

// MinimockFinish checks that all mocked methods have been called the expected number of times
func (m *WriteCloserMock) MinimockFinish() {
	m.finishOnce.Do(func() {
		if !m.minimockDone() {
			m.MinimockCloseInspect()

			m.MinimockWriteInspect()
			m.t.FailNow()
		}
	})
}

// MinimockWait waits for all mocked methods to be called the expected number of times
func (m *WriteCloserMock) MinimockWait(timeout mm_time.Duration) {
	timeoutCh := mm_time.After(timeout)
	for {
		if m.minimockDone() {
			return
		}
		select {
		case <-timeoutCh:
			m.MinimockFinish()
			return
		case <-mm_time.After(10 * mm_time.Millisecond):
		}
	}
}

func (m *WriteCloserMock) minimockDone() bool {
	done := true
	return done &&
		m.MinimockCloseDone() &&
		m.MinimockWriteDone()
}
