package di_test

import (
	"testing"

	"github.com/titosilva/drmchain-pos/internal/di"
)

type STestInterface interface {
	GetContent() int
	SetContent(int)
}

type STest struct {
	content int
}

type STest2 struct {
	content int
}

func (s STest2) GetContent() int {
	return s.content
}

func (s *STest2) SetContent(v int) {
	s.content = v
}

func Test__DISingleton__ShouldReturn__SameInstanceAlways(t *testing.T) {
	ctx := di.NewContext()
	di.AddSingleton[STest](ctx, func(*di.DIContext) *STest {
		return new(STest)
	})

	test := di.GetService[STest](ctx)
	test.content = 212031

	test2 := di.GetService[STest](ctx)

	if test.content != test2.content {
		t.Error("Different instances provided for STest")
	}
}

func Test__DIFactory__ShouldReturn__NewInstanceEveryTime(t *testing.T) {
	ctx := di.NewContext()
	di.AddFactory[STest2](ctx, func(*di.DIContext) *STest2 {
		return new(STest2)
	})

	st2 := di.GetService[STest2](ctx)
	st2.content = 1000

	st3 := di.GetService[STest2](ctx)
	st3.content = 2000

	if st2.content == st3.content {
		t.Error("Same instance provided for STest2")
	}

	if st2.content != 1000 {
		t.Error("st2 content has changed")
	}
}

func Test__DIInterfaceSingleton__ShouldReturn__NewInstanceEveryTime(t *testing.T) {
	ctx := di.NewContext()
	di.AddInterfaceSingleton[STestInterface](ctx, func(*di.DIContext) STestInterface {
		return new(STest2)
	})

	test := di.GetInterfaceService[STestInterface](ctx)
	test.SetContent(212031)

	test2 := di.GetInterfaceService[STestInterface](ctx)

	if test.GetContent() != test2.GetContent() {
		t.Error("Different instances provided for STest")
	}
}

func Test__DIInterfaceFactory__ShouldReturn__NewInstanceEveryTime(t *testing.T) {
	ctx := di.NewContext()
	di.AddInterfaceFactory[STestInterface](ctx, func(*di.DIContext) STestInterface {
		return new(STest2)
	})

	st2 := di.GetInterfaceService[STestInterface](ctx)
	st2.SetContent(1000)

	st3 := di.GetInterfaceService[STestInterface](ctx)
	st3.SetContent(2000)

	if st2.GetContent() == st3.GetContent() {
		t.Error("Same instance provided for STest2")
	}

	if st2.GetContent() != 1000 {
		t.Error("st2 content has changed")
	}
}
