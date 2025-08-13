package common

import "runtime"

// GetOperationName returns PackageName.FunctionName
//
// it should be used for logging or error wrapping
func GetOperationName() string {

	function, _, _, ok := runtime.Caller(1)
	if !ok {
		return "couldnt get op name"
	}

	return runtime.FuncForPC(function).Name()
}
