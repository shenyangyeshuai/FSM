package main

import (
	"fmt"
	"reflect"
)

type State interface {
	// 获取状态的名字
	Name() string

	// 该状态是否允许同状态转移
	EnableSameTransit() bool

	// 响应状态开始时
	OnBegin()

	// 响应状态结束时
	OnEnd()

	// 判断能够转移到某个状态
	CanTransitTo(name string) bool
}

func StateName(s State) string {
	if s == nil {
		return "none"
	}
	// 获取 指针s 指向的对象的名称
	return reflect.TypeOf(s).Elem().Name()
}

// 一个默认的实现, 提供基础行为
type StateInfo struct {
	name string
}

// 状态名
func (s *StateInfo) Name() string {
	return s.name
}

// 提供给内部设置名字
func (s *StateInfo) setName(name string) {
	s.name = name
}

// 不允许同状态转移
func (s *StateInfo) EnableSameTransit() bool {
	return false
}

// 默认状态开启时的实现
func (s *StateInfo) OnBegin() {
}

// 默认状态关闭时的实现
func (s *StateInfo) OnEnd() {
}

// 默认可以转移到任何状态
func (s *StateInfo) CanTransitTo(name string) bool {
	return true
}

type StateManager struct {
	// 已经添加的状态
	stateByName map[string]State

	// 状态改变时的回调
	OnChange func(from, to State)

	// 当前状态
	curr State
}

func (sm *StateManager) Add(s State) {
	name := StateName(s)

	// 将 s 转换为能设置名字的接口, 然后调用该接口
	s.(interface {
		setName(name string)
	}).setName(name)

	if sm.Get(name) != nil {
		panic("duplicate state: " + name)
	}

	sm.stateByName[name] = s
}

func (sm *StateManager) Get(name string) State {
	if v, ok := sm.stateByName[name]; ok {
		return v
	}

	return nil
}

func NewStateManager() *StateManager {
	return &StateManager{
		stateByName: make(map[string]State),
	}
}

// 状态未找到
var ErrStateNotFound = fmt.Errorf("state not found.")

// 禁止在同状态间转移
var ErrForbidSameStateTransit = fmt.Errorf("forbid same state transit.")

// 不能转移到指定状态
var ErrCannotTransitToState = fmt.Errorf("cannot transit to state.")

// 获取当前状态
func (sm *StateManager) CurrState() State {
	return sm.curr
}

// 当前状态能够转移到目标状态
func (sm *StateManager) CanCurrTransitTo(name string) bool {
	if sm.curr == nil {
		return true
	}

	if sm.curr.Name() == name && !sm.curr.EnableSameTransit() {
		return false
	}

	// 使用当前状态
	return sm.curr.CanTransitTo(name)
}

func (sm *StateManager) Transit(name string) error {
	next := sm.Get(name)
	if next == nil {
		return ErrStateNotFound
	}

	pre := sm.curr
	if sm.curr != nil {
		if sm.curr.Name() == name && !sm.curr.EnableSameTransit() {
			return ErrForbidSameStateTransit
		}

		if !sm.curr.CanTransitTo(name) {
			return ErrCannotTransitToState
		}

		sm.curr.OnEnd()
	}

	sm.curr = next
	sm.curr.OnBegin()

	if sm.OnChange != nil {
		sm.OnChange(pre, sm.curr)
	}

	return nil
}

type IdleState struct {
	StateInfo
}

func (i *IdleState) OnBegin() {
	fmt.Println("IdleState begin")
}

func (i *IdleState) OnEnd() {
	fmt.Println("IdleState end")
}

type MoveState struct {
	StateInfo
}

func (i *MoveState) OnBegin() {
	fmt.Println("MoveState begin")
}

func (i *MoveState) OnEnd() {
	fmt.Println("MoveState end")
}

func (i *MoveState) EnableSameTransit() bool {
	return true
}

type JumpState struct {
	StateInfo
}

func (i *JumpState) OnBegin() {
	fmt.Println("JumpState begin")
}

func (i *JumpState) OnEnd() {
	fmt.Println("JumpState end")
}

func (i *JumpState) CanTransitTo(name string) bool {
	return name != "MoveState"
}

func main() {
	sm := NewStateManager()
	sm.OnChange = func(from, to State) {
		fmt.Printf("%s ---> %s\n", StateName(from), StateName(to))
	}

	sm.Add(new(IdleState))
	sm.Add(new(MoveState))
	sm.Add(new(JumpState))

	transitAndReport(sm, "IdleState")
	transitAndReport(sm, "MoveState")
	transitAndReport(sm, "MoveState")
	transitAndReport(sm, "JumpState")
	transitAndReport(sm, "JumpState")
	transitAndReport(sm, "IdleState")
}

func transitAndReport(sm *StateManager, target string) {
	if err := sm.Transit(target); err != nil {
		fmt.Printf("Failed! %s --- > %s, %s\n\n", sm.CurrState().Name(), target, err.Error())
	}
}
