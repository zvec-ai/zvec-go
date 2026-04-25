# ZVec Go SDK

[English](README.md) | 中文

zvec-go 是 [zvec](https://github.com/alibaba/zvec) 向量数据库的 Go 语言 SDK，通过 cgo 封装 zvec C-API 实现。

## 简介

zvec 是一个高性能向量数据库，支持多种索引类型（HNSW、IVF、Flat、Invert）和丰富的数据类型。zvec-go 提供了完整的 Go 语言绑定，让您可以在 Go 项目中轻松使用 zvec 的强大功能。

## 环境要求

- **Go** ≥ 1.21
- **C 编译器**（gcc 或 clang），用于 cgo
- **CMake** ≥ 3.20 和 **Ninja**（用于构建 C-API 库）

## 快速开始

```bash
# 克隆仓库（包含子模块）
git clone --recursive https://github.com/zvec-ai/zvec-go.git
cd zvec-go

# 使用 Makefile 构建 C-API 库
make build-zvec

# 运行测试
make test
```

或者使用完整的构建命令：

```bash
# 克隆仓库（包含子模块）
git clone --recursive https://github.com/zvec-ai/zvec-go.git
cd zvec-go

# 从子模块构建 C-API 库
cd zvec && mkdir -p build && cd build
cmake .. -DCMAKE_BUILD_TYPE=Release -DBUILD_C_BINDINGS=ON -G Ninja
cmake --build . -j$(nproc 2>/dev/null || sysctl -n hw.ncpu) --target zvec_c_api
cd ../..

# 运行测试
go test -tags integration -count=1 -v ./...
```

## 安装

zvec-go 提供**两种构建模式**，适合不同的用户场景：

### 模式 1：Vendor 模式（默认 — `go get` + `go generate`）

预编译库通过 GitHub Releases 分发。使用 `go get` 获取代码，然后用 `go generate` 下载当前平台的预编译库：

```bash
# 1. 添加依赖
go get github.com/zvec-ai/zvec-go

# 2. 下载当前平台的预编译库
#    （从 GitHub Releases 下载，解压到 lib/ 目录）
go generate github.com/zvec-ai/zvec-go

# 3. 构建（需要 cgo）
CGO_ENABLED=1 go build .
```

支持平台：**Linux (x64, ARM64)**、**macOS (ARM64)** 和 **Windows (x64)**。

也可以指定版本：

```bash
go run github.com/zvec-ai/zvec-go/cmd/download-libs@latest -version v0.3.1
```

### 模式 2：Source 模式（从源码构建）

适合需要使用自定义 zvec 版本、参与项目开发或为不支持的平台构建的用户：

```bash
# 克隆仓库（包含子模块）
git clone --recursive https://github.com/zvec-ai/zvec-go.git
cd zvec-go

# 构建 C-API 库
make build-zvec

# 在您的项目中使用 replace 指令
# 在您项目的 go.mod 中：
#   require github.com/zvec-ai/zvec-go v0.0.0
#   replace github.com/zvec-ai/zvec-go => /path/to/zvec-go

# 使用 source 标签构建
CGO_ENABLED=1 go build -tags source ./...

# 运行测试
go test -tags "source integration" -v ./...
```

### 如何选择？

| 场景 | 模式 | 构建标签 |
|------|------|----------|
| 只想在项目中使用 zvec-go | **Vendor**（默认） | _（无需指定）_ |
| 参与 zvec-go 开发 | **Source** | `-tags source` |
| 需要自定义/最新版 zvec | **Source** | `-tags source` |
| 为不支持的平台构建 | **Source** | `-tags source` |
| AI/LLM Agent 集成 zvec-go | **Vendor**（默认） | _（无需指定）_ |

## 基本用法

```go
package main

import (
    "fmt"
    "log"

    zvec "github.com/zvec-ai/zvec-go"
)

func main() {
    // 初始化 zvec
    if err := zvec.Initialize(nil); err != nil {
        log.Fatal(err)
    }
    defer zvec.Shutdown()

    // 创建集合 Schema
    schema := zvec.NewCollectionSchema("example")
    defer schema.Destroy()

    // 添加 ID 字段（主键，使用倒排索引）
    idField := zvec.NewFieldSchema("id", zvec.DataTypeString, false, 0)
    idField.SetIndexParams(zvec.NewInvertIndexParams(true, false))
    schema.AddField(idField)

    // 添加向量字段（使用 HNSW 索引）
    embField := zvec.NewFieldSchema("embedding", zvec.DataTypeVectorFP32, false, 4)
    embField.SetIndexParams(zvec.NewHNSWIndexParams(zvec.MetricTypeCosine, 16, 200))
    schema.AddField(embField)

    // 创建并打开集合
    collection, err := zvec.CreateAndOpen("./my_data", schema, nil)
    if err != nil {
        log.Fatal(err)
    }
    defer collection.Close()

    // 插入文档
    doc := zvec.NewDoc()
    doc.SetPK("doc1")
    doc.AddStringField("id", "doc1")
    doc.AddVectorFP32Field("embedding", []float32{0.1, 0.2, 0.3, 0.4})
    collection.Insert([]*zvec.Doc{doc})
    doc.Destroy()

    // 向量查询
    query := zvec.NewVectorQuery()
    query.SetFieldName("embedding")
    query.SetQueryVector([]float32{0.4, 0.3, 0.3, 0.1})
    query.SetTopK(10)

    results, _ := collection.Query(query)
    query.Destroy()
    defer zvec.FreeDocs(results)

    for _, r := range results {
        fmt.Printf("PK=%s Score=%.4f\n", r.GetPK(), r.GetScore())
    }
}
```

## API 概览

### 初始化与配置

| API | 说明 |
|-----|------|
| `Initialize(config)` | 初始化 zvec 库 |
| `Shutdown()` | 关闭 zvec 库，释放资源 |
| `IsInitialized()` | 检查是否已初始化 |
| `GetVersion()` | 获取版本字符串 |
| `GetVersionMajor()` | 获取主版本号 |
| `GetVersionMinor()` | 获取次版本号 |
| `GetVersionPatch()` | 获取补丁版本号 |
| `CheckVersion(major, minor, patch)` | 检查版本是否兼容 |

### Schema 与索引

| API | 说明 |
|-----|------|
| `NewCollectionSchema(name)` | 创建集合 Schema |
| `NewFieldSchema(name, dataType, nullable, dim)` | 创建字段 Schema |
| `NewHNSWIndexParams(metricType, M, efConstruction)` | 创建 HNSW 索引参数 |
| `NewIVFIndexParams(metricType, nlist, nIters, useSoar)` | 创建 IVF 索引参数 |
| `NewFlatIndexParams(metricType)` | 创建 Flat 索引参数 |
| `NewInvertIndexParams(enable, wildcard)` | 创建倒排索引参数 |
| `SetIndexParams(params)` | 设置字段索引参数 |

### 集合操作

| API | 说明 |
|-----|------|
| `CreateAndOpen(path, schema, options)` | 创建并打开集合 |
| `Open(path, options)` | 打开现有集合 |
| `Close()` | 关闭集合 |
| `Destroy(path)` | 销毁集合 |
| `Flush()` | 刷新数据到磁盘 |
| `Optimize()` | 优化集合 |
| `GetStats()` | 获取集合统计信息 |
| `GetSchema()` | 获取集合 Schema |
| `GetOptions()` | 获取集合选项 |
| `AddColumn(field)` | 添加列 |
| `DropColumn(fieldName)` | 删除列 |
| `AlterColumn(fieldName, field)` | 修改列 |
| `CreateIndex(fieldName, params)` | 创建索引 |
| `DropIndex(fieldName)` | 删除索引 |

### 文档操作

| API | 说明 |
|-----|------|
| `NewDoc()` | 创建新文档 |
| `Destroy()` | 销毁文档，释放资源 |
| `SetPK(pk)` | 设置主键 |
| `GetPK()` | 获取主键 |
| `GetDocID()` | 获取文档 ID |
| `AddStringField(name, value)` | 添加字符串字段 |
| `AddBoolField(name, value)` | 添加布尔字段 |
| `AddInt32Field(name, value)` | 添加 Int32 字段 |
| `AddInt64Field(name, value)` | 添加 Int64 字段 |
| `AddFloatField(name, value)` | 添加 Float 字段 |
| `AddDoubleField(name, value)` | 添加 Double 字段 |
| `AddVectorFP32Field(name, value)` | 添加 FP32 向量字段 |
| `SetFieldNull(name)` | 设置字段为 NULL |
| `RemoveField(name)` | 删除字段 |
| `HasField(name)` | 检查字段是否存在 |

### 写入操作

| API | 说明 |
|-----|------|
| `Insert(docs)` | 插入文档 |
| `Update(docs)` | 更新文档 |
| `Upsert(docs)` | 插入或更新文档 |
| `Delete(pks)` | 根据主键删除文档 |
| `DeleteByFilter(filter)` | 根据条件删除文档 |

### 查询操作

| API | 说明 |
|-----|------|
| `NewVectorQuery()` | 创建向量查询对象 |
| `SetFieldName(name)` | 设置查询字段名 |
| `SetQueryVector(vector)` | 设置查询向量 |
| `SetTopK(k)` | 设置返回结果数量 |
| `SetFilter(filter)` | 设置过滤条件 |
| `SetOutputFields(fields)` | 设置输出字段 |
| `SetIncludeVector(include)` | 是否包含向量数据 |
| `SetIncludeDocID(include)` | 是否包含文档 ID |
| `Query(query)` | 执行查询 |
| `GroupByVectorQuery(query)` | 分组向量查询 |
| `Fetch(pks)` | 根据主键获取文档 |
| `FreeDocs(docs)` | 释放查询结果内存 |

### 数据类型

| 类型 | 说明 |
|-----|------|
| `DataTypeString` | 字符串类型 |
| `DataTypeBool` | 布尔类型 |
| `DataTypeInt32` | 32 位整数 |
| `DataTypeInt64` | 64 位整数 |
| `DataTypeUint32` | 32 位无符号整数 |
| `DataTypeUint64` | 64 位无符号整数 |
| `DataTypeFloat` | 单精度浮点数 |
| `DataTypeDouble` | 双精度浮点数 |
| `DataTypeVectorFP32` | FP32 向量 |
| `DataTypeBinary` | 二进制数据 |
| `DataTypeArray` | 数组类型 |
| `DataTypeSparseVector` | 稀疏向量 |

### 索引类型与度量

| 类型 | 说明 |
|-----|------|
| `MetricTypeL2` | L2 距离 |
| `MetricTypeIP` | 内积 |
| `MetricTypeCosine` | 余弦相似度 |
| `MetricTypeMIPSL2` | MIPSL2 距离 |
| `QuantizeTypeFP16` | FP16 量化 |
| `QuantizeTypeInt8` | Int8 量化 |
| `QuantizeTypeInt4` | Int4 量化 |

### 错误处理

| API | 说明 |
|-----|------|
| `Error.Code()` | 获取错误码 |
| `Error.Message()` | 获取错误信息 |
| `IsNotFound(err)` | 检查是否为"未找到"错误 |
| `IsAlreadyExists(err)` | 检查是否为"已存在"错误 |
| `IsInvalidArgument(err)` | 检查是否为"无效参数"错误 |

## 示例

项目提供了丰富的示例代码，帮助您快速上手：

- **examples/basic** — 基础用法示例，展示初始化、Schema 定义、CRUD 操作和向量查询
- **examples/schema_and_index** — Schema 与索引配置示例，展示如何定义不同的字段和索引类型
- **examples/crud_operations** — 完整的 CRUD 操作示例，包括插入、更新、删除等操作
- **examples/vector_query** — 向量查询示例，展示各种查询参数和过滤条件的使用
- **examples/collection_management** — 集合管理示例，展示集合的创建、打开、优化等操作
- **examples/error_handling** — 错误处理示例，展示如何正确处理各种错误情况
- **examples/configuration** — 全局配置示例，展示内存限制、线程数等配置选项的使用

运行示例：

```bash
cd examples/basic
go run main.go
```

## 开发指南

如果您想参与 zvec-go 的开发，请参考 [CONTRIBUTING.md](CONTRIBUTING.md) 了解详细的贡献指南。

## 同步 zvec 核心库

本仓库使用 **git submodule** 来跟踪 [zvec](https://github.com/alibaba/zvec) 核心库。更新方法：

```bash
# 更新到最新的 main 分支
./scripts/sync-zvec.sh

# 更新到指定的标签版本
./scripts/sync-zvec.sh v0.4.0
```

同时配置了 [Dependabot](https://docs.github.com/en/code-security/dependabot)，当 zvec submodule 有新的提交时会自动创建 PR。

## Makefile 命令

项目提供了便捷的 Makefile 命令来管理构建、测试和开发任务：

| 命令 | 说明 |
|-----|------|
| `make build-zvec` | 构建 zvec C-API 库 |
| `make build` | 构建 C-API 库并验证 Go 编译 |
| `make test` | 运行所有 Go 测试 |
| `make test-short` | 运行测试（短模式，跳过长时间运行的测试） |
| `make test-race` | 运行测试（带竞态检测器） |
| `make test-cover` | 运行测试并生成覆盖率报告 |
| `make bench` | 运行性能基准测试 |
| `make fuzz` | 运行模糊测试（默认每个目标 30 秒，可通过 `FUZZ_TIME` 自定义） |
| `make lint` | 运行所有 linter 检查 |
| `make vet` | 运行 go vet 检查 |
| `make fmt` | 格式化 Go 源文件 |
| `make fmt-check` | 检查 Go 文件格式（CI 友好） |
| `make sync-zvec` | 同步 zvec submodule 到最新 main 分支 |
| `make sync-zvec-build` | 同步 zvec submodule + 重新构建 + 测试 |
| `make check-zvec` | 检查上游 C-API 变更（不更新） |
| `make clean` | 清理构建产物 |
| `make deps` | 下载 Go 模块依赖 |
| `make install-tools` | 安装开发工具（golangci-lint、gofumpt） |
| `make all` | 执行完整的 CI 检查（构建、测试、lint） |
| `make help` | 显示帮助信息 |

## 支持平台

- Linux (x86_64, ARM64)
- macOS (ARM64)
- Windows (x86_64)

## 许可证

Apache License 2.0