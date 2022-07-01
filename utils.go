package core

import "strconv"

//获取整数类型的业务
func GenInt(data interface{}) int {
	var result int
	switch data.(type) {
	case uint:
		result = int(data.(uint))
		break
	case int8:
		result = int(data.(int8))
		break
	case uint8:
		result = int(data.(uint8))
		break
	case int16:
		result = int(data.(int16))
		break
	case uint16:
		result = int(data.(uint16))
		break
	case int32:
		result = int(data.(int32))
		break
	case uint32:
		result = int(data.(uint32))
		break
	case int64:
		result = int(data.(int64))
		break
	case uint64:
		result = int(data.(uint64))
		break
	case float32:
		result = int(data.(float32))
		break
	case float64:
		result = int(data.(float64))
		break
	case string:
		result, _ = strconv.Atoi(data.(string))
		break
	default:
		result = data.(int)
		break
	}
	return result
}
