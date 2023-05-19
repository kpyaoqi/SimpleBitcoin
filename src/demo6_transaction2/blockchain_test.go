package main

import (
	"fmt"
	"reflect"
	"strconv"
	"testing"
	"time"
)

func TestBase58Encode(t *testing.T) {
	block := &Block{time.Now().Unix(), nil, []byte{}, []byte{}, 0}
	serialize := block.Serialize()
	fmt.Println(serialize)
}

func TestByte(t *testing.T) {
	var b58Alphabet = []byte("123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz")
	for i, b := range b58Alphabet {
		fmt.Printf("%d:%x\n", i, b)
	}
}

func TestAddress(t *testing.T) {
	address := ValidateAddress("1B35qgq45pA5mpxFokzSA2nYEBAVTUWRgi")
	fmt.Sprintf(strconv.FormatBool(address))
}

func TestTime(t *testing.T) {
	timer := time.NewTimer(time.Second * 3) // 类型为 *time.Timer
	go func() {
		<-timer.C
		fmt.Println("timer 结束")
	}()

	time.Sleep(time.Second * 5)
	flag := timer.Stop() // 取消定时器
	fmt.Println(flag)    // false
}

func TestReflect(t *testing.T) {

	funcValue := reflect.ValueOf(add)
	params := []reflect.Value{reflect.ValueOf("lisi"), reflect.ValueOf(20)}

	reList := funcValue.Call(params)
	fmt.Println(reList) // 函数返回值

}
func add(name string, age int) {
	fmt.Printf("name is %s, age is %d \n", name, age)
}
