package generate

//go:generate buf generate

// 1. 通过buf管理依赖并生成go代码
// 2. 或者将依赖下载到本地third_party，通过 protoc -I. -I../third_party 方式指定依赖
