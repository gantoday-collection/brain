package utils

type Result struct {
	ResultCode    	string `json:"resultCode"`    //接口调用状态
	ResultMsg 		string `json:"resultMsg"` 		//接口调用说明
	ResultSubCode 	string`json:"resultSubCode"`	//接口反馈状态
	ResultSubMsg 	interface{}`json:"resultSubMsg"`	//接口反馈说明
}
