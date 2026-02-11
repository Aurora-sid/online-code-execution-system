@echo off
chcp 65001 >nul
echo ==========================================
echo 批量注册测试用户脚本
echo ==========================================
echo.

set BASE_URL=http://localhost:8080

echo 正在注册测试用户...
echo.

for /L %%i in (1,1,20) do (
    echo 注册用户: testuser%%i
    curl -s -X POST "%BASE_URL%/api/register" ^
        -H "Content-Type: application/json" ^
        -d "{\"username\": \"testuser%%i\", \"password\": \"password123\"}" > nul
)

echo.
echo ==========================================
echo 完成! 已创建 20 个测试用户
echo 用户名: testuser1 - testuser20
echo 密码: password123
echo ==========================================
pause
