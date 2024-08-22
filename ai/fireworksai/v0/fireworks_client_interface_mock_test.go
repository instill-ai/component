// Code generated by http://github.com/gojuno/minimock (v3.3.11). DO NOT EDIT.

package fireworksai

//go:generate minimock -i github.com/instill-ai/component/ai/fireworksai/v0.FireworksClientInterface -o fireworks_client_interface_mock_test.go -n FireworksClientInterfaceMock -p fireworksai

import (
	_ "embed"
	"sync"
	mm_atomic "sync/atomic"
	mm_time "time"

	"github.com/gojuno/minimock/v3"
)

// FireworksClientInterfaceMock implements FireworksClientInterface
type FireworksClientInterfaceMock struct {
	t          minimock.Tester
	finishOnce sync.Once

	funcChat          func(c1 ChatRequest) (c2 ChatResponse, err error)
	inspectFuncChat   func(c1 ChatRequest)
	afterChatCounter  uint64
	beforeChatCounter uint64
	ChatMock          mFireworksClientInterfaceMockChat

	funcEmbed          func(e1 EmbedRequest) (e2 EmbedResponse, err error)
	inspectFuncEmbed   func(e1 EmbedRequest)
	afterEmbedCounter  uint64
	beforeEmbedCounter uint64
	EmbedMock          mFireworksClientInterfaceMockEmbed
}

// NewFireworksClientInterfaceMock returns a mock for FireworksClientInterface
func NewFireworksClientInterfaceMock(t minimock.Tester) *FireworksClientInterfaceMock {
	m := &FireworksClientInterfaceMock{t: t}

	if controller, ok := t.(minimock.MockController); ok {
		controller.RegisterMocker(m)
	}

	m.ChatMock = mFireworksClientInterfaceMockChat{mock: m}
	m.ChatMock.callArgs = []*FireworksClientInterfaceMockChatParams{}

	m.EmbedMock = mFireworksClientInterfaceMockEmbed{mock: m}
	m.EmbedMock.callArgs = []*FireworksClientInterfaceMockEmbedParams{}

	t.Cleanup(m.MinimockFinish)

	return m
}

type mFireworksClientInterfaceMockChat struct {
	optional           bool
	mock               *FireworksClientInterfaceMock
	defaultExpectation *FireworksClientInterfaceMockChatExpectation
	expectations       []*FireworksClientInterfaceMockChatExpectation

	callArgs []*FireworksClientInterfaceMockChatParams
	mutex    sync.RWMutex

	expectedInvocations uint64
}

// FireworksClientInterfaceMockChatExpectation specifies expectation struct of the FireworksClientInterface.Chat
type FireworksClientInterfaceMockChatExpectation struct {
	mock      *FireworksClientInterfaceMock
	params    *FireworksClientInterfaceMockChatParams
	paramPtrs *FireworksClientInterfaceMockChatParamPtrs
	results   *FireworksClientInterfaceMockChatResults
	Counter   uint64
}

// FireworksClientInterfaceMockChatParams contains parameters of the FireworksClientInterface.Chat
type FireworksClientInterfaceMockChatParams struct {
	c1 ChatRequest
}

// FireworksClientInterfaceMockChatParamPtrs contains pointers to parameters of the FireworksClientInterface.Chat
type FireworksClientInterfaceMockChatParamPtrs struct {
	c1 *ChatRequest
}

// FireworksClientInterfaceMockChatResults contains results of the FireworksClientInterface.Chat
type FireworksClientInterfaceMockChatResults struct {
	c2  ChatResponse
	err error
}

// Marks this method to be optional. The default behavior of any method with Return() is '1 or more', meaning
// the test will fail minimock's automatic final call check if the mocked method was not called at least once.
// Optional() makes method check to work in '0 or more' mode.
// It is NOT RECOMMENDED to use this option by default unless you really need it, as it helps to
// catch the problems when the expected method call is totally skipped during test run.
func (mmChat *mFireworksClientInterfaceMockChat) Optional() *mFireworksClientInterfaceMockChat {
	mmChat.optional = true
	return mmChat
}

// Expect sets up expected params for FireworksClientInterface.Chat
func (mmChat *mFireworksClientInterfaceMockChat) Expect(c1 ChatRequest) *mFireworksClientInterfaceMockChat {
	if mmChat.mock.funcChat != nil {
		mmChat.mock.t.Fatalf("FireworksClientInterfaceMock.Chat mock is already set by Set")
	}

	if mmChat.defaultExpectation == nil {
		mmChat.defaultExpectation = &FireworksClientInterfaceMockChatExpectation{}
	}

	if mmChat.defaultExpectation.paramPtrs != nil {
		mmChat.mock.t.Fatalf("FireworksClientInterfaceMock.Chat mock is already set by ExpectParams functions")
	}

	mmChat.defaultExpectation.params = &FireworksClientInterfaceMockChatParams{c1}
	for _, e := range mmChat.expectations {
		if minimock.Equal(e.params, mmChat.defaultExpectation.params) {
			mmChat.mock.t.Fatalf("Expectation set by When has same params: %#v", *mmChat.defaultExpectation.params)
		}
	}

	return mmChat
}

// ExpectC1Param1 sets up expected param c1 for FireworksClientInterface.Chat
func (mmChat *mFireworksClientInterfaceMockChat) ExpectC1Param1(c1 ChatRequest) *mFireworksClientInterfaceMockChat {
	if mmChat.mock.funcChat != nil {
		mmChat.mock.t.Fatalf("FireworksClientInterfaceMock.Chat mock is already set by Set")
	}

	if mmChat.defaultExpectation == nil {
		mmChat.defaultExpectation = &FireworksClientInterfaceMockChatExpectation{}
	}

	if mmChat.defaultExpectation.params != nil {
		mmChat.mock.t.Fatalf("FireworksClientInterfaceMock.Chat mock is already set by Expect")
	}

	if mmChat.defaultExpectation.paramPtrs == nil {
		mmChat.defaultExpectation.paramPtrs = &FireworksClientInterfaceMockChatParamPtrs{}
	}
	mmChat.defaultExpectation.paramPtrs.c1 = &c1

	return mmChat
}

// Inspect accepts an inspector function that has same arguments as the FireworksClientInterface.Chat
func (mmChat *mFireworksClientInterfaceMockChat) Inspect(f func(c1 ChatRequest)) *mFireworksClientInterfaceMockChat {
	if mmChat.mock.inspectFuncChat != nil {
		mmChat.mock.t.Fatalf("Inspect function is already set for FireworksClientInterfaceMock.Chat")
	}

	mmChat.mock.inspectFuncChat = f

	return mmChat
}

// Return sets up results that will be returned by FireworksClientInterface.Chat
func (mmChat *mFireworksClientInterfaceMockChat) Return(c2 ChatResponse, err error) *FireworksClientInterfaceMock {
	if mmChat.mock.funcChat != nil {
		mmChat.mock.t.Fatalf("FireworksClientInterfaceMock.Chat mock is already set by Set")
	}

	if mmChat.defaultExpectation == nil {
		mmChat.defaultExpectation = &FireworksClientInterfaceMockChatExpectation{mock: mmChat.mock}
	}
	mmChat.defaultExpectation.results = &FireworksClientInterfaceMockChatResults{c2, err}
	return mmChat.mock
}

// Set uses given function f to mock the FireworksClientInterface.Chat method
func (mmChat *mFireworksClientInterfaceMockChat) Set(f func(c1 ChatRequest) (c2 ChatResponse, err error)) *FireworksClientInterfaceMock {
	if mmChat.defaultExpectation != nil {
		mmChat.mock.t.Fatalf("Default expectation is already set for the FireworksClientInterface.Chat method")
	}

	if len(mmChat.expectations) > 0 {
		mmChat.mock.t.Fatalf("Some expectations are already set for the FireworksClientInterface.Chat method")
	}

	mmChat.mock.funcChat = f
	return mmChat.mock
}

// When sets expectation for the FireworksClientInterface.Chat which will trigger the result defined by the following
// Then helper
func (mmChat *mFireworksClientInterfaceMockChat) When(c1 ChatRequest) *FireworksClientInterfaceMockChatExpectation {
	if mmChat.mock.funcChat != nil {
		mmChat.mock.t.Fatalf("FireworksClientInterfaceMock.Chat mock is already set by Set")
	}

	expectation := &FireworksClientInterfaceMockChatExpectation{
		mock:   mmChat.mock,
		params: &FireworksClientInterfaceMockChatParams{c1},
	}
	mmChat.expectations = append(mmChat.expectations, expectation)
	return expectation
}

// Then sets up FireworksClientInterface.Chat return parameters for the expectation previously defined by the When method
func (e *FireworksClientInterfaceMockChatExpectation) Then(c2 ChatResponse, err error) *FireworksClientInterfaceMock {
	e.results = &FireworksClientInterfaceMockChatResults{c2, err}
	return e.mock
}

// Times sets number of times FireworksClientInterface.Chat should be invoked
func (mmChat *mFireworksClientInterfaceMockChat) Times(n uint64) *mFireworksClientInterfaceMockChat {
	if n == 0 {
		mmChat.mock.t.Fatalf("Times of FireworksClientInterfaceMock.Chat mock can not be zero")
	}
	mm_atomic.StoreUint64(&mmChat.expectedInvocations, n)
	return mmChat
}

func (mmChat *mFireworksClientInterfaceMockChat) invocationsDone() bool {
	if len(mmChat.expectations) == 0 && mmChat.defaultExpectation == nil && mmChat.mock.funcChat == nil {
		return true
	}

	totalInvocations := mm_atomic.LoadUint64(&mmChat.mock.afterChatCounter)
	expectedInvocations := mm_atomic.LoadUint64(&mmChat.expectedInvocations)

	return totalInvocations > 0 && (expectedInvocations == 0 || expectedInvocations == totalInvocations)
}

// Chat implements FireworksClientInterface
func (mmChat *FireworksClientInterfaceMock) Chat(c1 ChatRequest) (c2 ChatResponse, err error) {
	mm_atomic.AddUint64(&mmChat.beforeChatCounter, 1)
	defer mm_atomic.AddUint64(&mmChat.afterChatCounter, 1)

	if mmChat.inspectFuncChat != nil {
		mmChat.inspectFuncChat(c1)
	}

	mm_params := FireworksClientInterfaceMockChatParams{c1}

	// Record call args
	mmChat.ChatMock.mutex.Lock()
	mmChat.ChatMock.callArgs = append(mmChat.ChatMock.callArgs, &mm_params)
	mmChat.ChatMock.mutex.Unlock()

	for _, e := range mmChat.ChatMock.expectations {
		if minimock.Equal(*e.params, mm_params) {
			mm_atomic.AddUint64(&e.Counter, 1)
			return e.results.c2, e.results.err
		}
	}

	if mmChat.ChatMock.defaultExpectation != nil {
		mm_atomic.AddUint64(&mmChat.ChatMock.defaultExpectation.Counter, 1)
		mm_want := mmChat.ChatMock.defaultExpectation.params
		mm_want_ptrs := mmChat.ChatMock.defaultExpectation.paramPtrs

		mm_got := FireworksClientInterfaceMockChatParams{c1}

		if mm_want_ptrs != nil {

			if mm_want_ptrs.c1 != nil && !minimock.Equal(*mm_want_ptrs.c1, mm_got.c1) {
				mmChat.t.Errorf("FireworksClientInterfaceMock.Chat got unexpected parameter c1, want: %#v, got: %#v%s\n", *mm_want_ptrs.c1, mm_got.c1, minimock.Diff(*mm_want_ptrs.c1, mm_got.c1))
			}

		} else if mm_want != nil && !minimock.Equal(*mm_want, mm_got) {
			mmChat.t.Errorf("FireworksClientInterfaceMock.Chat got unexpected parameters, want: %#v, got: %#v%s\n", *mm_want, mm_got, minimock.Diff(*mm_want, mm_got))
		}

		mm_results := mmChat.ChatMock.defaultExpectation.results
		if mm_results == nil {
			mmChat.t.Fatal("No results are set for the FireworksClientInterfaceMock.Chat")
		}
		return (*mm_results).c2, (*mm_results).err
	}
	if mmChat.funcChat != nil {
		return mmChat.funcChat(c1)
	}
	mmChat.t.Fatalf("Unexpected call to FireworksClientInterfaceMock.Chat. %v", c1)
	return
}

// ChatAfterCounter returns a count of finished FireworksClientInterfaceMock.Chat invocations
func (mmChat *FireworksClientInterfaceMock) ChatAfterCounter() uint64 {
	return mm_atomic.LoadUint64(&mmChat.afterChatCounter)
}

// ChatBeforeCounter returns a count of FireworksClientInterfaceMock.Chat invocations
func (mmChat *FireworksClientInterfaceMock) ChatBeforeCounter() uint64 {
	return mm_atomic.LoadUint64(&mmChat.beforeChatCounter)
}

// Calls returns a list of arguments used in each call to FireworksClientInterfaceMock.Chat.
// The list is in the same order as the calls were made (i.e. recent calls have a higher index)
func (mmChat *mFireworksClientInterfaceMockChat) Calls() []*FireworksClientInterfaceMockChatParams {
	mmChat.mutex.RLock()

	argCopy := make([]*FireworksClientInterfaceMockChatParams, len(mmChat.callArgs))
	copy(argCopy, mmChat.callArgs)

	mmChat.mutex.RUnlock()

	return argCopy
}

// MinimockChatDone returns true if the count of the Chat invocations corresponds
// the number of defined expectations
func (m *FireworksClientInterfaceMock) MinimockChatDone() bool {
	if m.ChatMock.optional {
		// Optional methods provide '0 or more' call count restriction.
		return true
	}

	for _, e := range m.ChatMock.expectations {
		if mm_atomic.LoadUint64(&e.Counter) < 1 {
			return false
		}
	}

	return m.ChatMock.invocationsDone()
}

// MinimockChatInspect logs each unmet expectation
func (m *FireworksClientInterfaceMock) MinimockChatInspect() {
	for _, e := range m.ChatMock.expectations {
		if mm_atomic.LoadUint64(&e.Counter) < 1 {
			m.t.Errorf("Expected call to FireworksClientInterfaceMock.Chat with params: %#v", *e.params)
		}
	}

	afterChatCounter := mm_atomic.LoadUint64(&m.afterChatCounter)
	// if default expectation was set then invocations count should be greater than zero
	if m.ChatMock.defaultExpectation != nil && afterChatCounter < 1 {
		if m.ChatMock.defaultExpectation.params == nil {
			m.t.Error("Expected call to FireworksClientInterfaceMock.Chat")
		} else {
			m.t.Errorf("Expected call to FireworksClientInterfaceMock.Chat with params: %#v", *m.ChatMock.defaultExpectation.params)
		}
	}
	// if func was set then invocations count should be greater than zero
	if m.funcChat != nil && afterChatCounter < 1 {
		m.t.Error("Expected call to FireworksClientInterfaceMock.Chat")
	}

	if !m.ChatMock.invocationsDone() && afterChatCounter > 0 {
		m.t.Errorf("Expected %d calls to FireworksClientInterfaceMock.Chat but found %d calls",
			mm_atomic.LoadUint64(&m.ChatMock.expectedInvocations), afterChatCounter)
	}
}

type mFireworksClientInterfaceMockEmbed struct {
	optional           bool
	mock               *FireworksClientInterfaceMock
	defaultExpectation *FireworksClientInterfaceMockEmbedExpectation
	expectations       []*FireworksClientInterfaceMockEmbedExpectation

	callArgs []*FireworksClientInterfaceMockEmbedParams
	mutex    sync.RWMutex

	expectedInvocations uint64
}

// FireworksClientInterfaceMockEmbedExpectation specifies expectation struct of the FireworksClientInterface.Embed
type FireworksClientInterfaceMockEmbedExpectation struct {
	mock      *FireworksClientInterfaceMock
	params    *FireworksClientInterfaceMockEmbedParams
	paramPtrs *FireworksClientInterfaceMockEmbedParamPtrs
	results   *FireworksClientInterfaceMockEmbedResults
	Counter   uint64
}

// FireworksClientInterfaceMockEmbedParams contains parameters of the FireworksClientInterface.Embed
type FireworksClientInterfaceMockEmbedParams struct {
	e1 EmbedRequest
}

// FireworksClientInterfaceMockEmbedParamPtrs contains pointers to parameters of the FireworksClientInterface.Embed
type FireworksClientInterfaceMockEmbedParamPtrs struct {
	e1 *EmbedRequest
}

// FireworksClientInterfaceMockEmbedResults contains results of the FireworksClientInterface.Embed
type FireworksClientInterfaceMockEmbedResults struct {
	e2  EmbedResponse
	err error
}

// Marks this method to be optional. The default behavior of any method with Return() is '1 or more', meaning
// the test will fail minimock's automatic final call check if the mocked method was not called at least once.
// Optional() makes method check to work in '0 or more' mode.
// It is NOT RECOMMENDED to use this option by default unless you really need it, as it helps to
// catch the problems when the expected method call is totally skipped during test run.
func (mmEmbed *mFireworksClientInterfaceMockEmbed) Optional() *mFireworksClientInterfaceMockEmbed {
	mmEmbed.optional = true
	return mmEmbed
}

// Expect sets up expected params for FireworksClientInterface.Embed
func (mmEmbed *mFireworksClientInterfaceMockEmbed) Expect(e1 EmbedRequest) *mFireworksClientInterfaceMockEmbed {
	if mmEmbed.mock.funcEmbed != nil {
		mmEmbed.mock.t.Fatalf("FireworksClientInterfaceMock.Embed mock is already set by Set")
	}

	if mmEmbed.defaultExpectation == nil {
		mmEmbed.defaultExpectation = &FireworksClientInterfaceMockEmbedExpectation{}
	}

	if mmEmbed.defaultExpectation.paramPtrs != nil {
		mmEmbed.mock.t.Fatalf("FireworksClientInterfaceMock.Embed mock is already set by ExpectParams functions")
	}

	mmEmbed.defaultExpectation.params = &FireworksClientInterfaceMockEmbedParams{e1}
	for _, e := range mmEmbed.expectations {
		if minimock.Equal(e.params, mmEmbed.defaultExpectation.params) {
			mmEmbed.mock.t.Fatalf("Expectation set by When has same params: %#v", *mmEmbed.defaultExpectation.params)
		}
	}

	return mmEmbed
}

// ExpectE1Param1 sets up expected param e1 for FireworksClientInterface.Embed
func (mmEmbed *mFireworksClientInterfaceMockEmbed) ExpectE1Param1(e1 EmbedRequest) *mFireworksClientInterfaceMockEmbed {
	if mmEmbed.mock.funcEmbed != nil {
		mmEmbed.mock.t.Fatalf("FireworksClientInterfaceMock.Embed mock is already set by Set")
	}

	if mmEmbed.defaultExpectation == nil {
		mmEmbed.defaultExpectation = &FireworksClientInterfaceMockEmbedExpectation{}
	}

	if mmEmbed.defaultExpectation.params != nil {
		mmEmbed.mock.t.Fatalf("FireworksClientInterfaceMock.Embed mock is already set by Expect")
	}

	if mmEmbed.defaultExpectation.paramPtrs == nil {
		mmEmbed.defaultExpectation.paramPtrs = &FireworksClientInterfaceMockEmbedParamPtrs{}
	}
	mmEmbed.defaultExpectation.paramPtrs.e1 = &e1

	return mmEmbed
}

// Inspect accepts an inspector function that has same arguments as the FireworksClientInterface.Embed
func (mmEmbed *mFireworksClientInterfaceMockEmbed) Inspect(f func(e1 EmbedRequest)) *mFireworksClientInterfaceMockEmbed {
	if mmEmbed.mock.inspectFuncEmbed != nil {
		mmEmbed.mock.t.Fatalf("Inspect function is already set for FireworksClientInterfaceMock.Embed")
	}

	mmEmbed.mock.inspectFuncEmbed = f

	return mmEmbed
}

// Return sets up results that will be returned by FireworksClientInterface.Embed
func (mmEmbed *mFireworksClientInterfaceMockEmbed) Return(e2 EmbedResponse, err error) *FireworksClientInterfaceMock {
	if mmEmbed.mock.funcEmbed != nil {
		mmEmbed.mock.t.Fatalf("FireworksClientInterfaceMock.Embed mock is already set by Set")
	}

	if mmEmbed.defaultExpectation == nil {
		mmEmbed.defaultExpectation = &FireworksClientInterfaceMockEmbedExpectation{mock: mmEmbed.mock}
	}
	mmEmbed.defaultExpectation.results = &FireworksClientInterfaceMockEmbedResults{e2, err}
	return mmEmbed.mock
}

// Set uses given function f to mock the FireworksClientInterface.Embed method
func (mmEmbed *mFireworksClientInterfaceMockEmbed) Set(f func(e1 EmbedRequest) (e2 EmbedResponse, err error)) *FireworksClientInterfaceMock {
	if mmEmbed.defaultExpectation != nil {
		mmEmbed.mock.t.Fatalf("Default expectation is already set for the FireworksClientInterface.Embed method")
	}

	if len(mmEmbed.expectations) > 0 {
		mmEmbed.mock.t.Fatalf("Some expectations are already set for the FireworksClientInterface.Embed method")
	}

	mmEmbed.mock.funcEmbed = f
	return mmEmbed.mock
}

// When sets expectation for the FireworksClientInterface.Embed which will trigger the result defined by the following
// Then helper
func (mmEmbed *mFireworksClientInterfaceMockEmbed) When(e1 EmbedRequest) *FireworksClientInterfaceMockEmbedExpectation {
	if mmEmbed.mock.funcEmbed != nil {
		mmEmbed.mock.t.Fatalf("FireworksClientInterfaceMock.Embed mock is already set by Set")
	}

	expectation := &FireworksClientInterfaceMockEmbedExpectation{
		mock:   mmEmbed.mock,
		params: &FireworksClientInterfaceMockEmbedParams{e1},
	}
	mmEmbed.expectations = append(mmEmbed.expectations, expectation)
	return expectation
}

// Then sets up FireworksClientInterface.Embed return parameters for the expectation previously defined by the When method
func (e *FireworksClientInterfaceMockEmbedExpectation) Then(e2 EmbedResponse, err error) *FireworksClientInterfaceMock {
	e.results = &FireworksClientInterfaceMockEmbedResults{e2, err}
	return e.mock
}

// Times sets number of times FireworksClientInterface.Embed should be invoked
func (mmEmbed *mFireworksClientInterfaceMockEmbed) Times(n uint64) *mFireworksClientInterfaceMockEmbed {
	if n == 0 {
		mmEmbed.mock.t.Fatalf("Times of FireworksClientInterfaceMock.Embed mock can not be zero")
	}
	mm_atomic.StoreUint64(&mmEmbed.expectedInvocations, n)
	return mmEmbed
}

func (mmEmbed *mFireworksClientInterfaceMockEmbed) invocationsDone() bool {
	if len(mmEmbed.expectations) == 0 && mmEmbed.defaultExpectation == nil && mmEmbed.mock.funcEmbed == nil {
		return true
	}

	totalInvocations := mm_atomic.LoadUint64(&mmEmbed.mock.afterEmbedCounter)
	expectedInvocations := mm_atomic.LoadUint64(&mmEmbed.expectedInvocations)

	return totalInvocations > 0 && (expectedInvocations == 0 || expectedInvocations == totalInvocations)
}

// Embed implements FireworksClientInterface
func (mmEmbed *FireworksClientInterfaceMock) Embed(e1 EmbedRequest) (e2 EmbedResponse, err error) {
	mm_atomic.AddUint64(&mmEmbed.beforeEmbedCounter, 1)
	defer mm_atomic.AddUint64(&mmEmbed.afterEmbedCounter, 1)

	if mmEmbed.inspectFuncEmbed != nil {
		mmEmbed.inspectFuncEmbed(e1)
	}

	mm_params := FireworksClientInterfaceMockEmbedParams{e1}

	// Record call args
	mmEmbed.EmbedMock.mutex.Lock()
	mmEmbed.EmbedMock.callArgs = append(mmEmbed.EmbedMock.callArgs, &mm_params)
	mmEmbed.EmbedMock.mutex.Unlock()

	for _, e := range mmEmbed.EmbedMock.expectations {
		if minimock.Equal(*e.params, mm_params) {
			mm_atomic.AddUint64(&e.Counter, 1)
			return e.results.e2, e.results.err
		}
	}

	if mmEmbed.EmbedMock.defaultExpectation != nil {
		mm_atomic.AddUint64(&mmEmbed.EmbedMock.defaultExpectation.Counter, 1)
		mm_want := mmEmbed.EmbedMock.defaultExpectation.params
		mm_want_ptrs := mmEmbed.EmbedMock.defaultExpectation.paramPtrs

		mm_got := FireworksClientInterfaceMockEmbedParams{e1}

		if mm_want_ptrs != nil {

			if mm_want_ptrs.e1 != nil && !minimock.Equal(*mm_want_ptrs.e1, mm_got.e1) {
				mmEmbed.t.Errorf("FireworksClientInterfaceMock.Embed got unexpected parameter e1, want: %#v, got: %#v%s\n", *mm_want_ptrs.e1, mm_got.e1, minimock.Diff(*mm_want_ptrs.e1, mm_got.e1))
			}

		} else if mm_want != nil && !minimock.Equal(*mm_want, mm_got) {
			mmEmbed.t.Errorf("FireworksClientInterfaceMock.Embed got unexpected parameters, want: %#v, got: %#v%s\n", *mm_want, mm_got, minimock.Diff(*mm_want, mm_got))
		}

		mm_results := mmEmbed.EmbedMock.defaultExpectation.results
		if mm_results == nil {
			mmEmbed.t.Fatal("No results are set for the FireworksClientInterfaceMock.Embed")
		}
		return (*mm_results).e2, (*mm_results).err
	}
	if mmEmbed.funcEmbed != nil {
		return mmEmbed.funcEmbed(e1)
	}
	mmEmbed.t.Fatalf("Unexpected call to FireworksClientInterfaceMock.Embed. %v", e1)
	return
}

// EmbedAfterCounter returns a count of finished FireworksClientInterfaceMock.Embed invocations
func (mmEmbed *FireworksClientInterfaceMock) EmbedAfterCounter() uint64 {
	return mm_atomic.LoadUint64(&mmEmbed.afterEmbedCounter)
}

// EmbedBeforeCounter returns a count of FireworksClientInterfaceMock.Embed invocations
func (mmEmbed *FireworksClientInterfaceMock) EmbedBeforeCounter() uint64 {
	return mm_atomic.LoadUint64(&mmEmbed.beforeEmbedCounter)
}

// Calls returns a list of arguments used in each call to FireworksClientInterfaceMock.Embed.
// The list is in the same order as the calls were made (i.e. recent calls have a higher index)
func (mmEmbed *mFireworksClientInterfaceMockEmbed) Calls() []*FireworksClientInterfaceMockEmbedParams {
	mmEmbed.mutex.RLock()

	argCopy := make([]*FireworksClientInterfaceMockEmbedParams, len(mmEmbed.callArgs))
	copy(argCopy, mmEmbed.callArgs)

	mmEmbed.mutex.RUnlock()

	return argCopy
}

// MinimockEmbedDone returns true if the count of the Embed invocations corresponds
// the number of defined expectations
func (m *FireworksClientInterfaceMock) MinimockEmbedDone() bool {
	if m.EmbedMock.optional {
		// Optional methods provide '0 or more' call count restriction.
		return true
	}

	for _, e := range m.EmbedMock.expectations {
		if mm_atomic.LoadUint64(&e.Counter) < 1 {
			return false
		}
	}

	return m.EmbedMock.invocationsDone()
}

// MinimockEmbedInspect logs each unmet expectation
func (m *FireworksClientInterfaceMock) MinimockEmbedInspect() {
	for _, e := range m.EmbedMock.expectations {
		if mm_atomic.LoadUint64(&e.Counter) < 1 {
			m.t.Errorf("Expected call to FireworksClientInterfaceMock.Embed with params: %#v", *e.params)
		}
	}

	afterEmbedCounter := mm_atomic.LoadUint64(&m.afterEmbedCounter)
	// if default expectation was set then invocations count should be greater than zero
	if m.EmbedMock.defaultExpectation != nil && afterEmbedCounter < 1 {
		if m.EmbedMock.defaultExpectation.params == nil {
			m.t.Error("Expected call to FireworksClientInterfaceMock.Embed")
		} else {
			m.t.Errorf("Expected call to FireworksClientInterfaceMock.Embed with params: %#v", *m.EmbedMock.defaultExpectation.params)
		}
	}
	// if func was set then invocations count should be greater than zero
	if m.funcEmbed != nil && afterEmbedCounter < 1 {
		m.t.Error("Expected call to FireworksClientInterfaceMock.Embed")
	}

	if !m.EmbedMock.invocationsDone() && afterEmbedCounter > 0 {
		m.t.Errorf("Expected %d calls to FireworksClientInterfaceMock.Embed but found %d calls",
			mm_atomic.LoadUint64(&m.EmbedMock.expectedInvocations), afterEmbedCounter)
	}
}

// MinimockFinish checks that all mocked methods have been called the expected number of times
func (m *FireworksClientInterfaceMock) MinimockFinish() {
	m.finishOnce.Do(func() {
		if !m.minimockDone() {
			m.MinimockChatInspect()

			m.MinimockEmbedInspect()
			m.t.FailNow()
		}
	})
}

// MinimockWait waits for all mocked methods to be called the expected number of times
func (m *FireworksClientInterfaceMock) MinimockWait(timeout mm_time.Duration) {
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

func (m *FireworksClientInterfaceMock) minimockDone() bool {
	done := true
	return done &&
		m.MinimockChatDone() &&
		m.MinimockEmbedDone()
}
