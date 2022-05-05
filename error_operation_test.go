package errGo

import (
	"1-19After/integralManagement/scs_api/app/smartcampus/models"
	"errors"
	"fmt"
	"testing"
)

func new2() (errNew error) {
	errNew = New(ErrTest)
	return
}

func new1() error {
	return new2()
}

func new11() (errNew error) {
	errNew = errors.New(ErrorsTest)

	return
}

func TestNew(t *testing.T) {
	var errNew error

	index := 1
	if index == 1 {
		errNew = new1()
		errNew = Wrap(errNew, "最外层")
		//addFCByIF(errNew)
	} else {
		errNew = new11()
	}

	if errNew != nil {
		fmt.Printf("%+v\n", errNew)
		fmt.Println("++++++++++++++")
		fmt.Println(errNew)
		fmt.Println("++++++++++++++")
		fmt.Println(Cause(errNew))
		fmt.Println("++++++++++++++")
		fmt.Println(errNew.Error())
	}
}

// 这个主要用于实际项目，需要放到框架中，这里可以看看逻辑。(运行不了)
func TestControl(t *testing.T) {

	err := Service(3)
	// 打印测试
	fmt.Printf("%+v\n", err)
	fmt.Println("++++++++++++++")
	fmt.Println(err)
	fmt.Println("++++++++++++++")
	fmt.Println(Cause(err))
	// 返回数据到前端
	if err != nil {
		//	// 集中对错误进行判断
		handleError(err)
		return
	}
	//app.ResponseSuccess()

}

// 一个模块（文件）写一个
func handleError(err error) {

	//status := HandleError(err)

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			//app.ResponseWithError(c, app.CodeNoUser, status)
		} else if errors.Is(err, models.FailedStudentHasExisted) {
			//app.ResponseWithError(c, app.CodeFailedStudentHasExisted, status)
		}
		//zap.L().Error("InsertAppraisalStudent failed", zap.Error(err))
		//app.ResponseWithError(c, app.CodeSelectOperationFail, status)
	}
	return
}

func Service(id int) error {

	// 两种写法
	//return DaoWrapf(id)

	//return DaoNew(id)

	return DaoWrapfBetter(id)
}

// 低级写法
func DaoWrapf(id int) (err error) {
	// 假设这是gorm库封装好的调用原生标准库中errors功能函数得到的一个错误
	err = errors.New(ErrQuery)
	err = Wrapf(err, "error getting the result with id: %d", id)
	return
}

// 进阶用法(伪代码)
func DaoWrapfBetter(id int) (err error) {
	// 这是gorm库封装好的调用原生标准库中errors功能函数得到的一个错误
	err = errors.New(ErrQuery)
	// 判断错误属于那种类型，资源未找到，代码错误，还是前端传参——对应三种错误码（Status Code）
	// 对应三种写法：
	err = BadRequest.Wrap(err, "测试")
	err = BadRequest.New("测试")
	err = NotFound.Wrap(err, "测试")
	err = NotFound.New("测试")
	err = NoType.Wrap(err, "测试")
	err = NoType.Newf("测试%d", 3)
	return

}

func DaoNew(id int) (err error) {
	// 使用函数Newf可以格式化输入。
	errNew := errors.New(ErrorsTest)
	err = Wrap(errNew, "id为3出错")
	return
}