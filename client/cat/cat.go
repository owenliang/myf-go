package cat

import (
	"github.com/owenliang/myf-go-cat/message"
)

// 为了构造嵌套事务，需要维护一个cat transaction栈
type MyfCat struct {
	stack []message.Transactor
	FinalStatus string
}

func NewMyfCat() (myfCat *MyfCat) {
	myfCat = &MyfCat{
		stack: make([]message.Transactor, 0),
	}
	return
}

// 追加transaction, 会自动挂载到父transaction
func (myfCat *MyfCat) Append(tran message.Transactor) {
	if len(myfCat.stack) != 0 {
		myfCat.stack[len(myfCat.stack)-1].AddChild(tran)
	}
	myfCat.stack = append(myfCat.stack, tran)
}

// 弹出transaction，会自动闭合transaction
func (myfCat *MyfCat) Pop() (tran message.Transactor) {
	if len(myfCat.stack) == 0 {
		return
	}
	tran = myfCat.stack[len(myfCat.stack)-1]
	if len(tran.GetStatus()) != 0 && len(myfCat.FinalStatus) == 0 {	// 需要将内层transaction错误保存下来
		myfCat.FinalStatus = tran.GetStatus()
	}
	tran.Complete()
	myfCat.stack = myfCat.stack[:len(myfCat.stack)-1]
	return
}

// 获取最近一层transaction
func (myfCat *MyfCat) Top() (tran message.Transactor) {
	if len(myfCat.stack) == 0 {
		return
	}
	tran = myfCat.stack[len(myfCat.stack)-1]
	return
}

// 获取transaction深度
func (myfCat *MyfCat) Depth() int {
	return len(myfCat.stack)
}

// 结束所有遗留事务
func (myfCat *MyfCat) FinishAll() {
	for {
		if myfCat.Depth() == 1 {	// root transaction
			if len(myfCat.FinalStatus) != 0 {	// 将嵌套transaction的错误传播到root transaction
				myfCat.Top().SetStatus(myfCat.FinalStatus)
			}
		} else if myfCat.Depth() == 0 {
			break
		}
		myfCat.Pop()
	}
}