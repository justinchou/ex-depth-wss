package service

import "strings"

// Observer 观察者模式接口
type Observer interface {
	Notify()
}

// UniqByLoop 通过两重循环过滤重复元素
func UniqByLoop(slc []string) []string {
	result := []string{} // 存放结果
	for _, v := range slc {
		v = strings.Trim(v, " ")
		if v == "" {
			continue
		}

		flag := true
		for j := range result {
			if v == result[j] {
				flag = false // 存在重复元素，标识为false
				break
			}
		}
		if flag { // 标识为false，不添加进结果
			result = append(result, v)
		}
	}
	return result
}

// UniqByMap 通过map主键唯一的特性过滤重复元素
func UniqByMap(slc []string) []string {
	result := []string{}
	tempMap := map[string]byte{} // 存放不重复主键
	for _, v := range slc {
		v = strings.Trim(v, " ")
		if v == "" {
			continue
		}

		l := len(tempMap)
		tempMap[v] = 0
		if len(tempMap) != l { // 加入map后，map长度变化，则元素不重复
			result = append(result, v)
		}
	}
	return result
}

// Uniq 元素去重
func Uniq(slc []string) []string {
	if len(slc) < 1024 {
		// 切片长度小于1024的时候，循环来过滤
		return UniqByLoop(slc)
	}

	// 大于的时候，通过map来过滤
	return UniqByMap(slc)
}

// Contains 数组中是否包含指定元素
func Contains(arr []string, str string) (contain bool) {
	contain = false
	for _, s := range arr {
		if s == str {
			contain = true
			break
		}
	}

	return contain
}
