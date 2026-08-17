# 本质评测环境说明

## 项目

- 项目编号：`zzz-gowork-0002`
- 项目名称：水质采样链路
- 项目说明：水质采样链路的标准库单进程业务服务

## 固定环境

- Go toolchain：`go1.26.5`
- go.mod language version：`go 1.21`
- GOTOOLCHAIN：`local`
- 支持平台：`linux/amd64`、`linux/arm64`
- Docker 基础镜像：`golang:1.26.5-bookworm`
- Docker manifest：`golang@sha256:53eeac89074db483fdf0ab3be1df32bf6e47562263d2d0d6baa7f26acb4957dd`

## 构建

```bash
./build_benzhi_docker.sh zzz-gowork-0002:benzhi-amd64 linux/amd64
./build_benzhi_docker.sh zzz-gowork-0002:benzhi-arm64 linux/arm64
```

## 运行

```bash
docker run --rm -it --network none zzz-gowork-0002:benzhi-amd64 bash
```

## 容器内验证

```bash
go version
go env GOTOOLCHAIN GOPROXY GOMODCACHE GOCACHE
go test ./...
go vet ./...
go build ./...
```

---

# 项目 README 同步内容

# 水质采样链路服务

本项目是一个水质采样链路管理服务，支持采样点、采样批次、样本瓶、交接记录和检测结论的完整管理。业务依次经历创建批次、完成采样、交接样本、确认结论，确保样本瓶归属当前批次且交接链完整后才能确认检测结论。

## 功能

- 创建和管理采样点
- 创建和管理采样批次
- 添加样本瓶至批次
- 完成采样并更新状态
- 记录样本交接链
- 确认检测结论
- 汇总待处理样本（待交接、待结论）
- 提供 Web 界面直接操作核心业务

## HTTP API

| Method | Path | 说明 |
|--------|------|------|
| POST | /api/points | 创建采样点 |
| GET | /api/points | 获取采样点列表 |
| GET | /api/points/{id} | 获取单个采样点 |
| POST | /api/batches | 创建采样批次 |
| GET | /api/batches | 获取批次列表 |
| GET | /api/batches/{id} | 获取批次详情 |
| POST | /api/batches/{id}/bottles | 向批次添加样本瓶 |
| POST | /api/batches/{id}/sampling | 完成采样 |
| POST | /api/batches/{id}/handover | 交接单个样本瓶 |
| POST | /api/batches/{id}/conclusion | 确认检测结论 |
| GET | /api/batches/{id}/samples | 获取批次下样本瓶 |
| GET | /api/samples/summary | 待处理样本汇总 |

## 运行

```bash
go run ./cmd/server
```

服务默认监听 `:8080`，打开浏览器访问 `/` 即可使用 Web 界面。

## 持久化

数据保存于本地 JSON 文件 `data.json`，采用安全写入（临时文件+原子重命名）方式，支持跨进程重启保留数据。

## 测试

```bash
go test ./...
```
