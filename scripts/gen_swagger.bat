@echo off
echo Starting Swagger generation...

REM %~dp0 代表脚本所在的当前目录
REM pushd 命令切换目录到脚本的上级目录（即项目根目录 FaceTaxLedger）
pushd "%~dp0.."

REM 检查是否安装了 swag
where swag >nul 2>nul
if %errorlevel% neq 0 (
    echo Swag tool not found. Installing...
    go install github.com/swaggo/swag/cmd/swag@latest
)

REM 执行生成命令
REM -g 指定入口文件（相对于根目录的路径）
REM -o 指定输出目录
REM --parseDependency --parseInternal 解析外部依赖和内部包（因为你有 internal 目录）
swag init -g cmd/server/main.go -o docs --parseInternal

REM 恢复目录
popd

echo Swagger generation complete.
pause