package errGo

import (
	"fmt"
	"io"
	"path"
	//"project/utils/config"
	"reflect"
	"runtime"
	"strconv"
	"strings"
)

// stack represents a stack of program counters.
type Stack []uintptr

// Frame represents a program counter inside a stack frame.
// For historical reasons if Frame is interpreted as a uintptr
// its value represents the program counter + 1.
type Frame uintptr

// pc returns the program counter for this frame;
// multiple frames may have the same PC value.
func (f Frame) pc() uintptr { return uintptr(f) - 1 }

// file returns the full path to the file that contains the
// function for this Frame`s pc
func (f Frame) file() string {
	fn := runtime.FuncForPC(f.pc()) //FuncForPC函数返回一个表示调用栈标识符pc对应的调用栈的*Func；每一个调用栈必然是对某个函数的调用；如果该调用栈标识符没有对应的调用栈，函数会返回nil。
	if fn == nil {
		return "unknown"
	}
	file, _ := fn.FileLine(f.pc()) //FileLine方法返回该调用栈所调用的函数的源代码文件名和行号。如果pc不是f内的调用栈标识符，结果是不精确的。
	return file
}

// line returns the line number of source code of the
// function for this Frame`s pc
func (f Frame) line() int {
	fn := runtime.FuncForPC(f.pc())
	if fn == nil {
		return 0
	}
	_, line := fn.FileLine(f.pc()) //返回行号
	return line
}

// name returns the name of this function, if known
func (f Frame) name() string {
	fn := runtime.FuncForPC(f.pc())
	if fn == nil {
		return "unknown"
	}
	return fn.Name() //Name方法返回该调用栈所调用的函数的名字。
}

func (f Frame) Format(s fmt.State, verb rune) { //fmt.State是一个接口
	switch verb {
	case 's':
		switch {
		case s.Flag('+'):
			io.WriteString(s, f.name()) //
			io.WriteString(s, "\n\t")   //\n是换行符号,\t是制表符
			// TODO 生产环境和开发环境有不同的错误提示
			io.WriteString(s, profile(f.file()))
		default:
			io.WriteString(s, path.Base(f.file()))
		}
	case 'd':
		io.WriteString(s, strconv.Itoa(f.line())) //Itoa函数将数字转化成字符串
	case 'n':
		io.WriteString(s, funcname(f.name()))
	case 'v':
		f.Format(s, 's')
		io.WriteString(s, ":")
		f.Format(s, 'd')
	}
}

// TODO Optimization: stack frames are stored only for methods and
//  functions that call new or wrap methods and functions
// Number of layers to store call stack information.
func callers(layer int) *Stack {
	const depth = 32
	var pcs [depth]uintptr
	// skip： The number of frames skipped from the top of the stack
	// pc slice: The method call stack for the goroutine is passed in
	//n := runtime.Callers(4, pcs[:])
	//var st Stack = pcs[:n - 3]

	n := runtime.Callers(layer, pcs[:])
	var st Stack = pcs[0:n]
	return &st
}

// Add function information.
func addFCByIF(itf interface{}) {
	pc, _, _, ok := runtime.Caller(1)

	if !ok {
		return
	}
	value := reflect.ValueOf(itf)
	element := value.Elem().Field(1).Interface()
	fund := reflect.ValueOf(element)
	methodFunc := fund.MethodByName("ModifyPC")
	if !methodFunc.IsValid() {
		return
	}

	params := []reflect.Value{reflect.ValueOf(pc)}

	methodFunc.Call(params)

}

//// Add call stack information.
//func addSt(stack *Stack) *Stack {
//	// 0: Represents information about the current method
//
//	return stack
//}

func (s *Stack) Format(st fmt.State, verb rune) {
	switch verb {
	case 'v':
		switch {
		case st.Flag('+'):
			for _, pc := range *s {
				f := Frame(pc)
				// use Fprintf to format and print to st
				fmt.Fprintf(st, "\n%+v", f)
			}
		}
	}
}

// funcname removes the path prefix component of a function`s name
// reported by func.Name().
func funcname(name string) string {
	i := strings.LastIndex(name, "/") //判断/在name中首次出现的位置，如没有出现则返回-1
	name = name[i+1:]
	i = strings.Index(name, ".")
	return name[i+1:]
}

// suitable for the project scenario, starting under the project directory.
func profile(name string) string {
	// If it is a Go library, it is processed by default.
	// TODO 下面这两个地方是哪种环境的意思。
	i := strings.Index(name, "go") //判断go在name中首次出现的位置，如没有出现则返回-1
	name = name[i:]
	/*if !strings.Contains(name, config.ApplicationConfig.ProName) {
			i := strings.Index(name, "go")
			name = name[i:]
		} else {
			i := strings.Index(name, config.ApplicationConfig.ProName)
			fmt.Println(i)
			name = name[i:]
	}*/
	return name
}
